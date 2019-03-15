package hcloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func dataSourceHcloudFloatingIP() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceHcloudFloatingIPRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"home_location": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"server_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"ip_address": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"ip_network": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Computed: true,
			},
			"selector": {
				Type:          schema.TypeString,
				Optional:      true,
				Deprecated:    "Please use the with_selector property instead.",
				ConflictsWith: []string{"with_selector"},
			},
			"with_selector": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"selector"},
			},
		},
	}
}

func dataSourceHcloudFloatingIPRead(d *schema.ResourceData, m interface{}) (err error) {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	var f *hcloud.FloatingIP
	if id, ok := d.GetOk("id"); ok {
		f, _, err = client.FloatingIP.GetByID(ctx, id.(int))
		if err != nil {
			return err
		}
		if f == nil {
			return fmt.Errorf("no Floating IP found with id %d", id)
		}
		setFloatingIPSchema(d, f)
		return
	}

	if ip, ok := d.GetOk("ip_address"); ok {
		var allIPs []*hcloud.FloatingIP
		allIPs, err = client.FloatingIP.All(ctx)
		if err != nil {
			return err
		}

		// Find by 'ip_address'
		for _, f := range allIPs {
			if f.IP.String() == ip.(string) {
				setFloatingIPSchema(d, f)
				return
			}
		}
		return fmt.Errorf("no Floating IP found with ip_address %s", ip)
	}

	var selector string
	if v := d.Get("with_selector").(string); v != "" {
		selector = v
	} else if v := d.Get("selector").(string); v != "" {
		selector = v
	}
	if selector != "" {
		var allIPs []*hcloud.FloatingIP
		opts := hcloud.FloatingIPListOpts{
			ListOpts: hcloud.ListOpts{
				LabelSelector: selector,
			},
		}
		allIPs, err = client.FloatingIP.AllWithOpts(ctx, opts)
		if err != nil {
			return err
		}
		if len(allIPs) == 0 {
			return fmt.Errorf("no Floating IP found for selector %q", selector)
		}
		if len(allIPs) > 1 {
			return fmt.Errorf("more than one Floating IP found for selector %q", selector)
		}
		setFloatingIPSchema(d, allIPs[0])
		return
	}

	return fmt.Errorf("please specify a id, ip_address or a selector to lookup the FloatingIP")
}
