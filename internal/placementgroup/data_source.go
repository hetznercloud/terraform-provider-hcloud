package placementgroup

import (
	"log"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/hcclient"
	"golang.org/x/net/context"
)

const DataSourceType = "hcloud_placement_group"

func DataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudPlacementGroupRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			"servers": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceHcloudPlacementGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)
	if id, ok := d.GetOk("id"); ok {
		i, _, err := client.PlacementGroup.GetByID(ctx, id.(int))
		if err != nil {
			return hcclient.ErrorToDiag(err)
		}
		if i == nil {
			return diag.Errorf("no placement group found with id %d", id)
		}
		setSchema(d, i)
		return nil
	}
	if name, ok := d.GetOk("name"); ok {
		i, _, err := client.PlacementGroup.GetByName(ctx, name.(string))
		if err != nil {
			return hcclient.ErrorToDiag(err)
		}
		if i == nil {
			return diag.Errorf("no placement group found with name %v", name)
		}
		setSchema(d, i)
		return nil
	}
	if selector, ok := d.GetOk("with_selector"); ok {
		var allPlacementGroups []*hcloud.PlacementGroup

		opts := hcloud.PlacementGroupListOpts{ListOpts: hcloud.ListOpts{LabelSelector: selector.(string)}}
		allPlacementGroups, err := client.PlacementGroup.AllWithOpts(ctx, opts)
		if err != nil {
			return hcclient.ErrorToDiag(err)
		}
		if len(allPlacementGroups) == 0 {
			return diag.Errorf("no placement group found for selector %q", selector)
		}
		if len(allPlacementGroups) == 1 {
			if _, ok := d.GetOk("most_recent"); !ok {
				return diag.Errorf("more than one placement group found for selector %q", selector)
			}
			sortPlacementGroupListByCreated(allPlacementGroups)
			log.Printf("[INFO] %d placement groups found for selector %q, using %d as the most recent one", len(allPlacementGroups), selector, allPlacementGroups[0].ID)
		}
		setSchema(d, allPlacementGroups[0])
		return nil
	}
	return diag.Errorf("please specify an id, a name or a selector to lookup the placement group")
}

func sortPlacementGroupListByCreated(placementGroupList []*hcloud.PlacementGroup) {
	sort.Slice(placementGroupList, func(i, j int) bool {
		return placementGroupList[i].Created.After(placementGroupList[j].Created)
	})
}
