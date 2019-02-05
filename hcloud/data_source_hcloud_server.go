package hcloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func dataSourceHcloudServer() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceHcloudServerRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"server_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"image": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"location": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"datacenter": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"backup_window": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"backups": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"ipv4_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipv6_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipv6_network": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"iso": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"rescue": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Computed: true,
			},
			"selector": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceHcloudServerRead(d *schema.ResourceData, m interface{}) (err error) {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	var s *hcloud.Server
	if id, ok := d.GetOk("id"); ok {
		s, _, err = client.Server.GetByID(ctx, id.(int))
		if err != nil {
			return err
		}
		if s == nil {
			return fmt.Errorf("no Server found with id %d", id)
		}
		setServerSchema(d, s)
		return
	}

	if name, ok := d.GetOk("name"); ok {
		s, _, err = client.Server.GetByName(ctx, name.(string))
		if err != nil {
			return err
		}
		if s == nil {
			return fmt.Errorf("no Server found with name %s", name)
		}
		setServerSchema(d, s)
		return
	}

	if selector, ok := d.GetOk("selector"); ok {
		var allServers []*hcloud.Server
		opts := hcloud.ServerListOpts{
			ListOpts: hcloud.ListOpts{
				LabelSelector: selector.(string),
			},
		}
		allServers, err = client.Server.AllWithOpts(ctx, opts)
		if err != nil {
			return err
		}
		if len(allServers) == 0 {
			return fmt.Errorf("no Server found for selector %q", selector)
		}
		if len(allServers) > 1 {
			return fmt.Errorf("more than one Server found for selector %q", selector)
		}
		setServerSchema(d, allServers[0])
		return
	}

	return fmt.Errorf("please specify a id, name or a selector to lookup the Server")
}
