package hcloud

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func dataSourceHcloudLocation() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudLocationRead,
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

func dataSourceHcloudLocationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	if id, ok := d.GetOk("id"); ok {
		l, _, err := client.Location.GetByID(ctx, id.(int))
		if err != nil {
			return diag.FromErr(err)
		}
		if l == nil {
			return diag.Errorf("no location found with id %d", id)
		}
		setLocationSchema(d, l)
		return nil
	}
	if name, ok := d.GetOk("name"); ok {
		l, _, err := client.Location.GetByName(ctx, name.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		if l == nil {
			return diag.Errorf("no location found with name %v", name)
		}
		setLocationSchema(d, l)
		return nil
	}

	return diag.Errorf("please specify an id, or a name to lookup for a location")
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
