package hcloud

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/firewall"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/hcclient"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/snapshot"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/certificate"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/datacenter"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/floatingip"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/image"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/loadbalancer"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/location"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/network"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/rdns"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/server"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/servertype"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/sshkey"
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
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("HCLOUD_TOKEN", nil),
				Description: "The API token to access the Hetzner cloud.",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					token := val.(string)
					if len(token) != 64 {
						errs = append(errs, errors.New("entered token is invalid (must be exactly 64 characters long)"))
					}
					return
				},
			},
			"endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("HCLOUD_ENDPOINT", nil),
			},
			"poll_interval": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			certificate.UploadedResourceType:  certificate.UploadedResource(),
			certificate.ResourceType:          certificate.UploadedResource(), // Alias for backwards compatibility.
			certificate.ManagedResourceType:   certificate.ManagedResource(),
			firewall.ResourceType:             firewall.Resource(),
			floatingip.AssignmentResourceType: floatingip.AssignmentResource(),
			floatingip.ResourceType:           floatingip.Resource(),
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
			sshkey.ResourceType:               sshkey.Resource(),
			volume.AttachmentResourceType:     volume.AttachmentResource(),
			volume.ResourceType:               volume.Resource(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			certificate.DataSourceType:           certificate.DataSource(),
			datacenter.DatacentersDataSourceType: datacenter.DatacentersDataSource(),
			datacenter.DataSourceType:            datacenter.DataSource(),
			firewall.DataSourceType:              firewall.DataSource(),
			floatingip.DataSourceType:            floatingip.DataSource(),
			image.DataSourceType:                 image.DataSource(),
			loadbalancer.DataSourceType:          loadbalancer.DataSource(),
			location.DataSourceType:              location.DataSource(),
			location.LocationsDataSourceType:     location.LocationsDataSource(),
			network.DataSourceType:               network.DataSource(),
			server.DataSourceType:                server.DataSource(),
			servertype.DataSourceType:            servertype.DataSource(),
			servertype.ServerTypesDataSourceType: servertype.ServerTypesDataSource(),
			sshkey.DataSourceType:                sshkey.DataSource(),
			sshkey.SSHKeysDataSourceType:         sshkey.SSHKeysDataSource(),
			volume.DataSourceType:                volume.DataSource(),
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
			return nil, hcclient.ErrorToDiag(err)
		}
		opts = append(opts, hcloud.WithPollInterval(pollInterval))
	}
	if logging.LogLevel() != "" {
		opts = append(opts, hcloud.WithDebugWriter(log.Writer()))
	}
	log.Printf("[DEBUG] hcloud terraform provider version: %s commit: %s", Version, Commit)
	log.Printf("[DEBUG] hcloud-go version: %s", hcloud.Version)
	return hcloud.NewClient(opts...), nil
}
