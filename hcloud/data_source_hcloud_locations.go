package hcloud

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func dataSourceHcloudLocations() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceHcloudLocationsRead,
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

func dataSourceHcloudLocationsRead(data *schema.ResourceData, m interface{}) (err error) {
	client := m.(*hcloud.Client)
	ctx := context.Background()
	locations, err := client.Location.All(ctx)
	if err != nil {
		return err
	}

	data.SetId(time.Now().UTC().String())
	var names, descriptions, ids []string
	for _, v := range locations {
		ids = append(ids, strconv.Itoa(v.ID))
		descriptions = append(descriptions, v.Description)
		names = append(names, v.Name)
	}
	data.Set("location_ids", ids)
	data.Set("names", names)
	data.Set("descriptions", descriptions)
	return
}
