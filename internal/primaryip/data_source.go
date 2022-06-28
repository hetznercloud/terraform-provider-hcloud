package primaryip

import (
	"context"
	"crypto/sha1"
	"fmt"
	"strconv"
	"strings"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/hcclient"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/datasourceutil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

const (
	// DataSourceType is the type name of the Hetzner Cloud Primary IP resource.
	DataSourceType = "hcloud_primary_ip"

	// DataSourceListType is the type name to receive a list of Hetzner Cloud Primary IPs resources.
	DataSourceListType = "hcloud_primary_ips"
)

// getCommonDataSchema returns a new common schema used by all primary ip data sources.
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
		"datacenter": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"assignee_id": {
			Type:     schema.TypeInt,
			Computed: true,
			Optional: true,
		},
		"assignee_type": {
			Type:     schema.TypeString,
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
		"auto_delete": {
			Type:     schema.TypeBool,
			Computed: true,
		},
	}
}

// DataSource creates a new Terraform schema for the hcloud_primary_ip data
// source.
func DataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudPrimaryIPRead,
		Schema: datasourceutil.MergeSchema(
			getCommonDataSchema(),
			map[string]*schema.Schema{
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
		ReadContext: dataSourceHcloudPrimaryIPListRead,
		Schema: map[string]*schema.Schema{
			"primary_ips": {
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

func dataSourceHcloudPrimaryIPRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	if id, ok := d.GetOk("id"); ok {
		f, _, err := client.PrimaryIP.GetByID(ctx, id.(int))
		if err != nil {
			return hcclient.ErrorToDiag(err)
		}
		if f == nil {
			return diag.Errorf("no Primary IP found with id %d", id)
		}
		setPrimaryIPSchema(d, f)
		return nil
	}
	if name, ok := d.GetOk("name"); ok {
		f, _, err := client.PrimaryIP.GetByName(ctx, name.(string))
		if err != nil {
			return hcclient.ErrorToDiag(err)
		}
		if f == nil {
			return diag.Errorf("no Primary IP found with name %s", name)
		}
		setPrimaryIPSchema(d, f)
		return nil
	}
	if ip, ok := d.GetOk("ip_address"); ok {
		primaryIP, _, err := client.PrimaryIP.GetByIP(ctx, ip.(string))
		if err != nil {
			return hcclient.ErrorToDiag(err)
		}
		setPrimaryIPSchema(d, primaryIP)
		return nil
	}

	var selector string
	if v, ok := d.Get("with_selector").(string); ok && v != "" {
		selector = v
	} else if v, ok := d.Get("selector").(string); ok && v != "" {
		selector = v
	}
	if selector != "" {
		var allIPs []*hcloud.PrimaryIP
		opts := hcloud.PrimaryIPListOpts{
			ListOpts: hcloud.ListOpts{
				LabelSelector: selector,
			},
		}
		allIPs, _, err := client.PrimaryIP.List(ctx, opts)
		if err != nil {
			return hcclient.ErrorToDiag(err)
		}
		if len(allIPs) == 0 {
			return diag.Errorf("no Primary IP found for selector %q", selector)
		}
		if len(allIPs) > 1 {
			return diag.Errorf("more than one Primary IP found for selector %q", selector)
		}
		setPrimaryIPSchema(d, allIPs[0])
		return nil
	}

	return diag.Errorf("please specify a id, ip_address or a selector to lookup the Primary IP")
}

func dataSourceHcloudPrimaryIPListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	selector := d.Get("with_selector").(string)

	var allIPs []*hcloud.PrimaryIP
	opts := hcloud.PrimaryIPListOpts{
		ListOpts: hcloud.ListOpts{
			LabelSelector: selector,
		},
	}
	allIPs, _, err := client.PrimaryIP.List(ctx, opts)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}

	ids := make([]string, len(allIPs))
	tfIPs := make([]map[string]interface{}, len(allIPs))
	for i, ip := range allIPs {
		ids[i] = strconv.Itoa(ip.ID)
		tfIPs[i] = getPrimaryIPAttributes(ip)
	}
	d.Set("primary_ips", tfIPs)
	d.SetId(fmt.Sprintf("%x", sha1.Sum([]byte(strings.Join(ids, "")))))

	return nil
}
