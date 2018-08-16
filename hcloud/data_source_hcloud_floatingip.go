package hcloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"strconv"
)

func dataSourceHcloudFloatingIP() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceHcloudFloatingIPRead,
		Schema: map[string]*schema.Schema{
			"ip_address": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"server_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"home_location": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}
func setFloatingIPSchema(d *schema.ResourceData, f *hcloud.FloatingIP) error {
	d.SetId(strconv.Itoa(f.ID))
	if err := d.Set("ip_address", f.IP.String()); err != nil {
		return err
	}
	if f.Server != nil {
		if err := d.Set("server_id", f.Server.ID); err != nil {
			return err
		}
	}
	if err := d.Set("type", f.Type); err != nil {
		return err
	}
	if err := d.Set("home_location", f.HomeLocation.Name); err != nil {
		return err
	}
	return nil
}
func dataSourceHcloudFloatingIPRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()
	allIPs, err := client.FloatingIP.All(ctx)
	if err != nil {
		return err
	}
	ip, ok := d.GetOk("ip_address")
	if !ok {
		return fmt.Errorf("please specify a floating ip")
	}
	// Find by 'ip_address'
	for _, f := range allIPs {
		if f.IP.String() == ip.(string) {
			return setFloatingIPSchema(d, f)
		}
	}
	return fmt.Errorf("could not find floating ip %s", ip)
}
