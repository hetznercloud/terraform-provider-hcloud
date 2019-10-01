package hcloud

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func dataSourceHcloudLocation() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceHcloudLocationRead,
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
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"country": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"city": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"latitude": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"longitude": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
		},
	}
}

func dataSourceHcloudLocationRead(d *schema.ResourceData, m interface{}) (err error) {
	client := m.(*hcloud.Client)
	ctx := context.Background()
	var l *hcloud.Location
	if id, ok := d.GetOk("id"); ok {
		l, _, err = client.Location.GetByID(ctx, id.(int))
		if err != nil {
			return err
		}
		if l == nil {
			return fmt.Errorf("no location found with id %d", id)
		}
		setLocationSchema(d, l)
		return
	}
	if name, ok := d.GetOk("name"); ok {
		l, _, err = client.Location.GetByName(ctx, name.(string))
		if err != nil {
			return err
		}
		if l == nil {
			return fmt.Errorf("no location found with name %v", name)
		}
		setLocationSchema(d, l)
		return
	}

	return fmt.Errorf("please specify an id, or a name to lookup for a location")
}

func setLocationSchema(d *schema.ResourceData, l *hcloud.Location) {
	d.SetId(strconv.Itoa(l.ID))
	d.Set("name", l.Name)
	d.Set("description", l.Description)
	d.Set("country", l.Country)
	d.Set("city", l.City)
	d.Set("latitude", l.Latitude)
	d.Set("longitude", l.Longitude)
}
