package hcloud

import (
	"errors"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
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
			"hcloud_server":                 resourceServer(),
			"hcloud_floating_ip":            resourceFloatingIP(),
			"hcloud_floating_ip_assignment": resourceFloatingIPAssignment(),
			"hcloud_ssh_key":                resourceSSHKey(),
			"hcloud_rdns":                   resourceReverseDNS(),
			"hcloud_volume":                 resourceVolume(),
			"hcloud_volume_attachment":      resourceVolumeAttachment(),
			"hcloud_network":                resourceNetwork(),
			"hcloud_network_subnet":         resourceNetworkSubnet(),
			"hcloud_network_route":          resourceNetworkRoute(),
			"hcloud_server_network":         resourceServerNetwork(),
			"hcloud_load_balancer":          resourceLoadBalancer(),
			"hcloud_load_balancer_service":  resourceLoadBalancerService(),
			"hcloud_load_balancer_network":  resourceLoadBalancerNetwork(),
			"hcloud_load_balancer_target":   resourceLoadBalancerTarget(),
			"hcloud_certificate":            resourceCertificate(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"hcloud_datacenter":    dataSourceHcloudDatacenter(),
			"hcloud_datacenters":   dataSourceHcloudDatacenters(),
			"hcloud_floating_ip":   dataSourceHcloudFloatingIP(),
			"hcloud_image":         dataSourceHcloudImage(),
			"hcloud_location":      dataSourceHcloudLocation(),
			"hcloud_locations":     dataSourceHcloudLocations(),
			"hcloud_server":        dataSourceHcloudServer(),
			"hcloud_server_type":   dataSourceHcloudServerType(),
			"hcloud_server_types":  dataSourceHcloudServerTypes(),
			"hcloud_ssh_key":       dataSourceHcloudSSHKey(),
			"hcloud_ssh_keys":      dataSourceHcloudSSHKeys(),
			"hcloud_volume":        dataSourceHcloudVolume(),
			"hcloud_network":       dataSourceHcloudNetwork(),
			"hcloud_load_balancer": dataSourceHcloudLoadBalancer(),
			"hcloud_certificate":   dataSourceHcloudCertificate(),
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
