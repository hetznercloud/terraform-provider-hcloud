package hcloud

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func dataSourceHcloudServerTypes() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudServerTypesRead,
		Schema: map[string]*schema.Schema{
			"server_type_ids": {
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

func dataSourceHcloudServerTypesRead(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)
	ds, err := client.ServerType.All(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(time.Now().UTC().String())
	var names, descriptions, ids []string
	for _, v := range ds {
		ids = append(ids, strconv.Itoa(v.ID))
		descriptions = append(descriptions, v.Description)
		names = append(names, v.Name)
	}
	data.Set("server_type_ids", ids)
	data.Set("names", names)
	data.Set("descriptions", descriptions)
	return nil
}
