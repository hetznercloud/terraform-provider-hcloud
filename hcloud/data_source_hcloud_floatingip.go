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

	return fmt.Errorf("please specify a id or ip_address to lookup the FloatingIP")
}
