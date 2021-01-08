package hcloud

import (
	"context"
	"crypto/sha1"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"strconv"
	"strings"
)

func dataSourceHcloudDatacenters() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudDatacentersRead,
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

func dataSourceHcloudDatacentersRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)
	dcs, err := client.Datacenter.All(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	var names, descriptions, ids []string
	for _, v := range dcs {
		ids = append(ids, strconv.Itoa(v.ID))
		descriptions = append(descriptions, v.Description)
		names = append(names, v.Name)
	}
	d.SetId(fmt.Sprintf("%x", sha1.Sum([]byte(strings.Join(ids, "")))))
	d.Set("datacenter_ids", ids)
	d.Set("names", names)
	d.Set("descriptions", descriptions)
	return nil
}
