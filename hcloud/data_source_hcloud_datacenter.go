package hcloud

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func dataSourceHcloudDatacenter() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudDatacenterRead,
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
			"location": {
				Type:     schema.TypeMap,
				Computed: true,
			},
			"supported_server_type_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},
			"available_server_type_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},
		},
	}
}

func dataSourceHcloudDatacenterRead(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	if id, ok := data.GetOk("id"); ok {
		d, _, err := client.Datacenter.GetByID(ctx, id.(int))
		if err != nil {
			return diag.FromErr(err)
		}
		if d == nil {
			return diag.Errorf("no datacenter found with id %d", id)
		}
		setDatacenterSchema(data, d)
		return nil
	}
	if name, ok := data.GetOk("name"); ok {
		d, _, err := client.Datacenter.GetByName(ctx, name.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		if d == nil {
			return diag.Errorf("no datacenter found with name %v", name)
		}
		setDatacenterSchema(data, d)
		return nil
	}

	return diag.Errorf("please specify an id, or a name to lookup for a datacenter")
}

func setDatacenterSchema(data *schema.ResourceData, d *hcloud.Datacenter) {
	data.SetId(strconv.Itoa(d.ID))
	data.Set("name", d.Name)
	data.Set("description", d.Description)
	data.Set("location", map[string]string{
		"id":          strconv.Itoa(d.Location.ID),
		"name":        d.Location.Name,
		"description": d.Location.Description,
		"country":     d.Location.Country,
		"city":        d.Location.City,
		"latitude":    fmt.Sprintf("%f", d.Location.Latitude),
		"longitude":   fmt.Sprintf("%f", d.Location.Longitude),
	})
	var supported, available []int
	for _, v := range d.ServerTypes.Supported {
		supported = append(supported, v.ID)
	}
	for _, v := range d.ServerTypes.Available {
		available = append(available, v.ID)
	}
	sort.Ints(available)
	sort.Ints(supported)
	data.Set("supported_server_type_ids", supported)
	data.Set("available_server_type_ids", available)
}
