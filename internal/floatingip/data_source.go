package floatingip

import (
	"context"
	"strconv"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/merge"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

const (
	// DataSourceType is the type name of the Hetzner Cloud Floating IP resource.
	DataSourceType = "hcloud_floating_ip"

	// DataSourceListType is the type name to receive a list of Hetzner Cloud Floating IPs resources.
	DataSourceListType = "hcloud_floating_ips"
)

// getCommonDataSchema returns a new common schema used by all floating ip data sources.
func getCommonDataSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
		},
		"name": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"type": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"description": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"home_location": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"server_id": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"ip_address": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"ip_network": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"labels": {
			Type:     schema.TypeMap,
			Computed: true,
		},
		"delete_protection": {
			Type:     schema.TypeBool,
			Computed: true,
		},
	}
}

// DataSource creates a new Terraform schema for the hcloud_floating_ip data
// source.
func DataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudFloatingIPRead,
		Schema: merge.Maps(
			getCommonDataSchema(),
			map[string]*schema.Schema{
				"selector": {
					Type:          schema.TypeString,
					Optional:      true,
					Deprecated:    "Please use the with_selector property instead.",
					ConflictsWith: []string{"with_selector"},
				},
				"with_selector": {
					Type:          schema.TypeString,
					Optional:      true,
					ConflictsWith: []string{"selector"},
				},
			},
		),
	}
}

func DataSourceList() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudFloatingIPListRead,
		Schema: map[string]*schema.Schema{
			"floating_ips": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: getCommonDataSchema(),
				},
			},
			"with_selector": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceHcloudFloatingIPRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	if id, ok := d.GetOk("id"); ok {
		f, _, err := client.FloatingIP.GetByID(ctx, id.(int))
		if err != nil {
			return hcloudutil.ErrorToDiag(err)
		}
		if f == nil {
			return diag.Errorf("no Floating IP found with id %d", id)
		}
		setFloatingIPSchema(d, f)
		return nil
	}
	if name, ok := d.GetOk("name"); ok {
		f, _, err := client.FloatingIP.GetByName(ctx, name.(string))
		if err != nil {
			return hcloudutil.ErrorToDiag(err)
		}
		if f == nil {
			return diag.Errorf("no Floating IP found with name %s", name)
		}
		setFloatingIPSchema(d, f)
		return nil
	}
	if ip, ok := d.GetOk("ip_address"); ok {
		var allIPs []*hcloud.FloatingIP
		allIPs, err := client.FloatingIP.All(ctx)
		if err != nil {
			return hcloudutil.ErrorToDiag(err)
		}

		// Find by 'ip_address'
		for _, f := range allIPs {
			if f.IP.String() == ip.(string) {
				setFloatingIPSchema(d, f)
				return nil
			}
		}
		return diag.Errorf("no Floating IP found with ip_address %s", ip)
	}

	var selector string
	if v := d.Get("with_selector").(string); v != "" {
		selector = v
	} else if v := d.Get("selector").(string); v != "" {
		selector = v
	}
	if selector != "" {
		var allIPs []*hcloud.FloatingIP
		opts := hcloud.FloatingIPListOpts{
			ListOpts: hcloud.ListOpts{
				LabelSelector: selector,
			},
		}
		allIPs, err := client.FloatingIP.AllWithOpts(ctx, opts)
		if err != nil {
			return hcloudutil.ErrorToDiag(err)
		}
		if len(allIPs) == 0 {
			return diag.Errorf("no Floating IP found for selector %q", selector)
		}
		if len(allIPs) > 1 {
			return diag.Errorf("more than one Floating IP found for selector %q", selector)
		}
		setFloatingIPSchema(d, allIPs[0])
		return nil
	}

	return diag.Errorf("please specify a id, ip_address or a selector to lookup the FloatingIP")
}

func dataSourceHcloudFloatingIPListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	selector := d.Get("with_selector").(string)

	var allIPs []*hcloud.FloatingIP
	opts := hcloud.FloatingIPListOpts{
		ListOpts: hcloud.ListOpts{
			LabelSelector: selector,
		},
	}
	allIPs, err := client.FloatingIP.AllWithOpts(ctx, opts)
	if err != nil {
		return hcloudutil.ErrorToDiag(err)
	}

	ids := make([]string, len(allIPs))
	tfIPs := make([]map[string]interface{}, len(allIPs))
	for i, ip := range allIPs {
		ids[i] = strconv.Itoa(ip.ID)
		tfIPs[i] = getFloatingIPAttributes(ip)
	}
	d.Set("floating_ips", tfIPs)
	d.SetId(datasourceutil.ListID(ids))

	return nil
}
