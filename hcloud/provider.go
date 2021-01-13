package hcloud

import (
	"errors"
	"log"
	"time"

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

var Version = "not build yet"
var Commit = "not build yet"

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
		},
		ResourcesMap: map[string]*schema.Resource{
			"hcloud_server":                 server.Resource(),
			"hcloud_floating_ip":            floatingip.Resource(),
			"hcloud_floating_ip_assignment": floatingip.AssignmentResource(),
			"hcloud_ssh_key":                sshkey.Resource(),
			"hcloud_rdns":                   rdns.Resource(),
			"hcloud_volume":                 volume.Resource(),
			"hcloud_volume_attachment":      volume.ResourceVolumeAttachment(),
			"hcloud_network":                network.Resource(),
			"hcloud_network_subnet":         network.SubnetResource(),
			"hcloud_network_route":          network.ResourceRoute(),
			"hcloud_server_network":         server.ResourceNetwork(),
			"hcloud_load_balancer":          loadbalancer.Resource(),
			"hcloud_load_balancer_service":  loadbalancer.ServiceResource(),
			"hcloud_load_balancer_network":  loadbalancer.NetworkResource(),
			"hcloud_load_balancer_target":   loadbalancer.TargetResource(),
			"hcloud_certificate":            certificate.Resource(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"hcloud_datacenter":    datacenter.DataSource(),
			"hcloud_datacenters":   datacenter.DataSourceDatacenters(),
			"hcloud_floating_ip":   floatingip.DataSource(),
			"hcloud_image":         image.DataSource(),
			"hcloud_location":      location.DataSource(),
			"hcloud_locations":     location.DataSourceLocations(),
			"hcloud_server":        server.DataSource(),
			"hcloud_server_type":   servertype.DataSource(),
			"hcloud_server_types":  servertype.DataSourceServerTypes(),
			"hcloud_ssh_key":       sshkey.DataSource(),
			"hcloud_ssh_keys":      sshkey.DataSourceSSHKeys(),
			"hcloud_volume":        volume.DataSource(),
			"hcloud_network":       network.DataSource(),
			"hcloud_load_balancer": loadbalancer.DataSource(),
			"hcloud_certificate":   certificate.DataSource(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
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
			return nil, err
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
