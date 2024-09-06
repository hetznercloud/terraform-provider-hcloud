package firewall

import (
	"context"
	"log"
	"sort"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/datasourceutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
)

const (
	// DataSourceType is the type name of the Hetzner Cloud Firewall resource.
	DataSourceType = "hcloud_firewall"

	// DataSourceListType is the type name to receive a list of Hetzner Cloud Firewall resources.
	DataSourceListType = "hcloud_firewalls"
)

// getCommonDataSchema returns a new common schema used by all firewall data sources.
func getCommonDataSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
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
		"apply_to": {
			Type:     schema.TypeSet,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"label_selector": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					"server": {
						Type:     schema.TypeInt,
						Optional: true,
						Computed: true,
					},
				},
			},
		},
		"rule": {
			Type:     schema.TypeSet,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"direction": {
						Type:     schema.TypeString,
						Required: true,
					},
					"protocol": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"port": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"source_ips": {
						Type: schema.TypeSet,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
						Optional: true,
					},
					"destination_ips": {
						Type: schema.TypeSet,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
						Optional: true,
					},
					"description": {
						Type:     schema.TypeString,
						Optional: true,
					},
				},
			},
		},
	}
}

func DataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudFirewallRead,
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
		ReadContext: dataSourceHcloudFirewallListRead,
		Schema: map[string]*schema.Schema{
			"firewalls": {
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

func dataSourceHcloudFirewallRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)
	if id, ok := d.GetOk("id"); ok {
		i, _, err := client.Firewall.GetByID(ctx, id.(int))
		if err != nil {
			return hcloudutil.ErrorToDiag(err)
		}
		if i == nil {
			return diag.Errorf("no firewall found with id %d", id)
		}
		setFirewallSchema(d, i)
		return nil
	}
	if name, ok := d.GetOk("name"); ok {
		i, _, err := client.Firewall.GetByName(ctx, name.(string))
		if err != nil {
			return hcloudutil.ErrorToDiag(err)
		}
		if i == nil {
			return diag.Errorf("no firewall found with name %v", name)
		}
		setFirewallSchema(d, i)
		return nil
	}

	if selector, ok := d.GetOk("with_selector"); ok {
		var allFirewalls []*hcloud.Firewall

		opts := hcloud.FirewallListOpts{ListOpts: hcloud.ListOpts{LabelSelector: selector.(string)}}
		allFirewalls, err := client.Firewall.AllWithOpts(ctx, opts)
		if err != nil {
			return hcloudutil.ErrorToDiag(err)
		}
		if len(allFirewalls) == 0 {
			return diag.Errorf("no firewall found for selector %q", selector)
		}
		if len(allFirewalls) > 1 {
			if _, ok := d.GetOk("most_recent"); !ok {
				return diag.Errorf("more than one firewall found for selector %q", selector)
			}
			sortFirewallListByCreated(allFirewalls)
			log.Printf("[INFO] %d firewalls found for selector %q, using %d as the most recent one", len(allFirewalls), selector, allFirewalls[0].ID)
		}
		setFirewallSchema(d, allFirewalls[0])
		return nil
	}
	return diag.Errorf("please specify an id, a name or a selector to lookup the firewall")
}

func dataSourceHcloudFirewallListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	selector := d.Get("with_selector").(string)

	opts := hcloud.FirewallListOpts{ListOpts: hcloud.ListOpts{LabelSelector: selector}}
	allFirewalls, err := client.Firewall.AllWithOpts(ctx, opts)
	if err != nil {
		return hcloudutil.ErrorToDiag(err)
	}

	if _, ok := d.GetOk("most_recent"); ok {
		sortFirewallListByCreated(allFirewalls)
	}

	ids := make([]string, len(allFirewalls))
	tfFirewalls := make([]map[string]interface{}, len(allFirewalls))
	for i, firewall := range allFirewalls {
		ids[i] = strconv.Itoa(firewall.ID)
		tfFirewalls[i] = getFirewallAttributes(firewall)
	}
	d.Set("firewalls", tfFirewalls)
	d.SetId(datasourceutil.ListID(ids))

	return nil
}

func sortFirewallListByCreated(firewallList []*hcloud.Firewall) {
	sort.Slice(firewallList, func(i, j int) bool {
		return firewallList[i].Created.After(firewallList[j].Created)
	})
}
