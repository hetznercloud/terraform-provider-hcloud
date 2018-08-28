package hcloud

import (
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

// Provider returns the hcloud terraform provider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("HCLOUD_TOKEN", nil),
				Description: "The API token to access the hetzner cloud.",
			},
			"endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("HCLOUD_ENDPOINT", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"hcloud_server":      resourceServer(),
			"hcloud_floating_ip": resourceFloatingIP(),
			"hcloud_ssh_key":     resourceSSHKey(),
			"hcloud_rdns":        resourceReverseDNS(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"hcloud_floating_ip": dataSourceHcloudFloatingIP(),
			"hcloud_ssh_key": dataSourceHcloudSSHKey(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	opts := []hcloud.ClientOption{
		hcloud.WithToken(d.Get("token").(string)),
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
	return hcloud.NewClient(opts...), nil
}
