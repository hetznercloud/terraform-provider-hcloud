package hcloud

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/certificate"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/firewall"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/floatingip"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/image"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/loadbalancer"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/network"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/placementgroup"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/primaryip"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/rdns"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/server"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/snapshot"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/volume"
)

// Build coordinates..
var (
	Version = "not build yet"
	Commit  = "not build yet"
)

// Provider returns the hcloud terraform provider.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("HCLOUD_TOKEN", nil),
				Description: "The Hetzner Cloud API token, can also be specified with the HCLOUD_TOKEN environment variable.",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) { // nolint:revive
					token := val.(string)
					if len(token) != 64 {
						errs = append(errs, errors.New("entered token is invalid (must be exactly 64 characters long)"))
					}
					return
				},
				Sensitive: true,
			},
			"endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("HCLOUD_ENDPOINT", nil),
				Description: "The Hetzner Cloud API endpoint, can be used to override the default API Endpoint https://api.hetzner.cloud/v1.",
			},
			"poll_interval": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "500ms",
				Description: "The interval at which actions are polled by the client. Default `500ms`. Increase this interval if you run into rate limiting errors.",
			},
			"poll_function": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "exponential",
				Description:  "The type of function to be used during the polling.",
				ValidateFunc: validation.StringInSlice([]string{"constant", "exponential"}, false),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			certificate.UploadedResourceType:  certificate.UploadedResource(),
			certificate.ResourceType:          certificate.UploadedResource(), // Alias for backwards compatibility.
			certificate.ManagedResourceType:   certificate.ManagedResource(),
			firewall.ResourceType:             firewall.Resource(),
			firewall.AttachmentResourceType:   firewall.AttachmentResource(),
			floatingip.AssignmentResourceType: floatingip.AssignmentResource(),
			floatingip.ResourceType:           floatingip.Resource(),
			primaryip.ResourceType:            primaryip.Resource(),
			loadbalancer.NetworkResourceType:  loadbalancer.NetworkResource(),
			loadbalancer.ResourceType:         loadbalancer.Resource(),
			loadbalancer.ServiceResourceType:  loadbalancer.ServiceResource(),
			loadbalancer.TargetResourceType:   loadbalancer.TargetResource(),
			network.ResourceType:              network.Resource(),
			network.RouteResourceType:         network.RouteResource(),
			network.SubnetResourceType:        network.SubnetResource(),
			rdns.ResourceType:                 rdns.Resource(),
			server.NetworkResourceType:        server.NetworkResource(),
			server.ResourceType:               server.Resource(),
			snapshot.ResourceType:             snapshot.Resource(),
			volume.AttachmentResourceType:     volume.AttachmentResource(),
			volume.ResourceType:               volume.Resource(),
			placementgroup.ResourceType:       placementgroup.Resource(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			certificate.DataSourceType:        certificate.DataSource(),
			certificate.DataSourceListType:    certificate.DataSourceList(),
			firewall.DataSourceType:           firewall.DataSource(),
			firewall.DataSourceListType:       firewall.DataSourceList(),
			floatingip.DataSourceType:         floatingip.DataSource(),
			floatingip.DataSourceListType:     floatingip.DataSourceList(),
			primaryip.DataSourceType:          primaryip.DataSource(),
			primaryip.DataSourceListType:      primaryip.DataSourceList(),
			image.DataSourceType:              image.DataSource(),
			image.DataSourceListType:          image.DataSourceList(),
			loadbalancer.DataSourceType:       loadbalancer.DataSource(),
			loadbalancer.DataSourceListType:   loadbalancer.DataSourceList(),
			network.DataSourceType:            network.DataSource(),
			network.DataSourceListType:        network.DataSourceList(),
			placementgroup.DataSourceType:     placementgroup.DataSource(),
			placementgroup.DataSourceListType: placementgroup.DataSourceList(),
			server.DataSourceType:             server.DataSource(),
			server.DataSourceListType:         server.DataSourceList(),
			volume.DataSourceType:             volume.DataSource(),
			volume.DataSourceListType:         volume.DataSourceList(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	opts := []hcloud.ClientOption{
		hcloud.WithToken(d.Get("token").(string)),
		hcloud.WithApplication("hcloud-terraform", Version),
	}
	if endpoint, ok := d.GetOk("endpoint"); ok {
		opts = append(opts, hcloud.WithEndpoint(endpoint.(string)))
	}
	if pollInterval, ok := d.GetOk("poll_interval"); ok {
		pollInterval, err := time.ParseDuration(pollInterval.(string))
		if err != nil {
			return nil, hcloudutil.ErrorToDiag(err)
		}
		pollFunction, ok := d.GetOk("poll_function")
		if ok && pollFunction == "constant" {
			opts = append(opts, hcloud.WithPollOpts(hcloud.PollOpts{BackoffFunc: hcloud.ConstantBackoff(pollInterval)}))
		} else {
			opts = append(opts, hcloud.WithPollOpts(hcloud.PollOpts{BackoffFunc: hcloud.ExponentialBackoff(2, pollInterval)}))
		}
	}
	if logging.LogLevel() != "" {
		opts = append(opts, hcloud.WithDebugWriter(log.Writer()))
	}
	log.Printf("[DEBUG] hcloud terraform provider version: %s commit: %s", Version, Commit)
	log.Printf("[DEBUG] hcloud-go version: %s", hcloud.Version)
	return hcloud.NewClient(opts...), nil
}
