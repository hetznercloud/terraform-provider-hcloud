package hcloud

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func dataSourceHcloudDatacenters() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceHcloudDatacentersRead,
		Schema: map[string]*schema.Schema{
			"datacenter_ids": {
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

func dataSourceHcloudDatacentersRead(data *schema.ResourceData, m interface{}) (err error) {
	client := m.(*hcloud.Client)
	ctx := context.Background()
	ds, err := client.Datacenter.All(ctx)
	if err != nil {
		return err
	}

	data.SetId(time.Now().UTC().String())
	var names, descriptions, ids []string
	for _, v := range ds {
		ids = append(ids, strconv.Itoa(v.ID))
		descriptions = append(descriptions, v.Description)
		names = append(names, v.Name)
	}
	data.Set("datacenter_ids", ids)
	data.Set("names", names)
	data.Set("descriptions", descriptions)
	return
}
