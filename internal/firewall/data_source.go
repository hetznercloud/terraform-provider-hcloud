package firewall

import (
	"context"
	"log"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/hcclient"
)

// DataSourceType is the type name of the Hetzner Cloud Firewall resource.
const DataSourceType = "hcloud_firewall"

func DataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudFirewallRead,
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
			"rule": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"direction": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"protocol": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"port": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"source_ips": &schema.Schema{
							Type: schema.TypeSet,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional: true,
						},
						"destination_ips": &schema.Schema{
							Type: schema.TypeSet,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional: true,
						},
					},
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
			return hcclient.ErrorToDiag(err)
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
			return hcclient.ErrorToDiag(err)
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
			return hcclient.ErrorToDiag(err)
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

func sortFirewallListByCreated(firewallList []*hcloud.Firewall) {
	sort.Slice(firewallList, func(i, j int) bool {
		return firewallList[i].Created.After(firewallList[j].Created)
	})
}
