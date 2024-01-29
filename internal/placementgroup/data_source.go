package placementgroup

import (
	"crypto/sha1"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.org/x/net/context"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/datasourceutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
)

const (
	// DataSourceType is the type name of the Hetzner Cloud Placement Group resource.
	DataSourceType = "hcloud_placement_group"

	// DataSourceListType is the type name to receive a list of Hetzner Cloud Placement Group resources.
	DataSourceListType = "hcloud_placement_groups"
)

// getCommonDataSchema returns a new common schema used by all placement group data sources.
func getCommonDataSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Type:     schema.TypeInt,
			Computed: true,
			Optional: true,
		},
		"name": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"labels": {
			Type:     schema.TypeMap,
			Computed: true,
			Optional: true,
		},
		"servers": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeInt,
			},
		},
		"type": {
			Type:     schema.TypeString,
			Computed: true,
			Optional: true,
		},
	}
}

func DataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudPlacementGroupRead,
		Schema: datasourceutil.MergeSchema(
			getCommonDataSchema(),
			map[string]*schema.Schema{
				"most_recent": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"with_selector": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
		),
	}
}

func DataSourceList() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudPlacementGroupListRead,
		Schema: map[string]*schema.Schema{
			"placement_groups": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: getCommonDataSchema(),
				},
			},
			"most_recent": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"with_selector": {
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
			return hcloudutil.ErrorToDiag(err)
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
			return hcloudutil.ErrorToDiag(err)
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
			return hcloudutil.ErrorToDiag(err)
		}
		if len(allPlacementGroups) == 0 {
			return diag.Errorf("no placement group found for selector %q", selector)
		}
		if len(allPlacementGroups) > 1 {
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

func dataSourceHcloudPlacementGroupListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	selector := d.Get("with_selector")

	opts := hcloud.PlacementGroupListOpts{ListOpts: hcloud.ListOpts{LabelSelector: selector.(string)}}
	allPlacementGroups, err := client.PlacementGroup.AllWithOpts(ctx, opts)
	if err != nil {
		return hcloudutil.ErrorToDiag(err)
	}

	if _, ok := d.GetOk("most_recent"); ok {
		sortPlacementGroupListByCreated(allPlacementGroups)
	}

	ids := make([]string, len(allPlacementGroups))
	tfPlacementGroups := make([]map[string]interface{}, len(allPlacementGroups))
	for i, firewall := range allPlacementGroups {
		ids[i] = strconv.Itoa(firewall.ID)
		tfPlacementGroups[i] = getAttributes(firewall)
	}
	d.Set("placement_groups", tfPlacementGroups)
	d.SetId(fmt.Sprintf("%x", sha1.Sum([]byte(strings.Join(ids, "")))))

	return nil
}

func sortPlacementGroupListByCreated(placementGroupList []*hcloud.PlacementGroup) {
	sort.Slice(placementGroupList, func(i, j int) bool {
		return placementGroupList[i].Created.After(placementGroupList[j].Created)
	})
}
