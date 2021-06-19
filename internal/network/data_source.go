package network

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/hcclient"
)

// DataSourceType is the type name of the Hetzner Cloud Network resource.
const DataSourceType = "hcloud_network"

// DataSource creates a new Terraform schema for the hcloud_network resource.
func DataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudNetworkRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ip_range": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			"with_selector": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}
func dataSourceHcloudNetworkRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	if id, ok := d.GetOk("id"); ok {
		n, _, err := client.Network.GetByID(ctx, id.(int))
		if err != nil {
			return hcclient.ErrorToDiag(err)
		}
		if n == nil {
			return diag.Errorf("no network found with id %d", id)
		}
		setNetworkSchema(d, n)
		return nil
	}
	if name, ok := d.GetOk("name"); ok {
		n, _, err := client.Network.GetByName(ctx, name.(string))
		if err != nil {
			return hcclient.ErrorToDiag(err)
		}
		if n == nil {
			return diag.Errorf("no network found with name %s", name)
		}
		setNetworkSchema(d, n)
		return nil
	}

	selector := d.Get("with_selector").(string)
	if selector != "" {
		var allNetworks []*hcloud.Network

		opts := hcloud.NetworkListOpts{
			ListOpts: hcloud.ListOpts{
				LabelSelector: selector,
			},
		}
		allNetworks, err := client.Network.AllWithOpts(ctx, opts)
		if err != nil {
			return hcclient.ErrorToDiag(err)
		}
		if len(allNetworks) == 0 {
			return diag.Errorf("no network found for selector %q", selector)
		}
		if len(allNetworks) > 1 {
			return diag.Errorf("more than one network found for selector %q", selector)
		}
		setNetworkSchema(d, allNetworks[0])
		return nil
	}
	return diag.Errorf("please specify an id, a name or a selector to lookup the network")
}
