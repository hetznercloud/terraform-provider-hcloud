package hcloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func dataSourceHcloudNetwork() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceHcloudNetworkRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ip_range": {
				Type:     schema.TypeString,
				Required: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			"with_selector": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"selector"},
			},
		},
	}
}
func dataSourceHcloudNetworkRead(d *schema.ResourceData, m interface{}) (err error) {
	client := m.(*hcloud.Client)
	ctx := context.Background()
	var n *hcloud.Network
	if id, ok := d.GetOk("id"); ok {
		n, _, err = client.Network.GetByID(ctx, id.(int))
		if err != nil {
			return err
		}
		if n == nil {
			return fmt.Errorf("no network found with id %d", id)
		}
		setNetworkSchema(d, n)
		return
	}
	if name, ok := d.GetOk("name"); ok {
		n, _, err = client.Network.GetByName(ctx, name.(string))
		if err != nil {
			return err
		}
		if n == nil {
			return fmt.Errorf("no network found with name %n", name)
		}
		setNetworkSchema(d, n)
		return
	}

	var selector string
	if v := d.Get("with_selector").(string); v != "" {
		selector = v
	} else if v := d.Get("selector").(string); v != "" {
		selector = v
	}
	if selector != "" {
		var allNetworks []*hcloud.Network

		opts := hcloud.NetworkListOpts{
			ListOpts: hcloud.ListOpts{
				LabelSelector: selector,
			},
		}
		allNetworks, err = client.Network.AllWithOpts(ctx, opts)
		if err != nil {
			return err
		}
		if len(allNetworks) == 0 {
			return fmt.Errorf("no network found for selector %q", selector)
		}
		if len(allNetworks) > 1 {
			return fmt.Errorf("more than one network found for selector %q", selector)
		}
		setNetworkSchema(d, allNetworks[0])
		return
	}
	return fmt.Errorf("please specify an id, a name or a selector to lookup the network")
}
