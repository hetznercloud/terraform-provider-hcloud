package hcloud

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func dataSourceHcloudLocations() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudLocationsRead,
		Schema: map[string]*schema.Schema{
			"location_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"names": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"descriptions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceHcloudLocationsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)
	locations, err := client.Location.All(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(time.Now().UTC().String())
	var names, descriptions, ids []string
	for _, v := range locations {
		ids = append(ids, strconv.Itoa(v.ID))
		descriptions = append(descriptions, v.Description)
		names = append(names, v.Name)
	}
	d.Set("location_ids", ids)
	d.Set("names", names)
	d.Set("descriptions", descriptions)
	return nil
}
