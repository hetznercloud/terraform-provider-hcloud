package network

import (
	"context"
	"crypto/sha1"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/datasourceutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
)

const (
	// DataSourceType is the type name of the Hetzner Cloud Network resource.
	DataSourceType = "hcloud_network"

	// DataSourceListType is the type name to receive a list of Hetzner Cloud Network resources.
	DataSourceListType = "hcloud_networks"
)

// getCommonDataSchema returns a new common schema used by all network data sources.
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
			Computed: true,
		},
		"ip_range": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"labels": {
			Type:     schema.TypeMap,
			Optional: true,
			Computed: true,
		},
		"delete_protection": {
			Type:     schema.TypeBool,
			Computed: true,
		},
		"expose_routes_to_vswitch": {
			Type:        schema.TypeBool,
			Description: "Indicates if the routes from this network should be exposed to the vSwitch connection. The exposing only takes effect if a vSwitch connection is active.",
			Computed:    true,
		},
	}
}

// DataSource creates a new Terraform schema for the hcloud_network resource.
func DataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudNetworkRead,
		Schema: datasourceutil.MergeSchema(
			getCommonDataSchema(),
			map[string]*schema.Schema{
				"most_recent": {
					Type:       schema.TypeBool,
					Optional:   true,
					Deprecated: "This attribute has no purpose.",
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
		ReadContext: dataSourceHcloudNetworkListRead,
		Schema: map[string]*schema.Schema{
			"networks": {
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

func dataSourceHcloudNetworkRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	if id, ok := d.GetOk("id"); ok {
		n, _, err := client.Network.GetByID(ctx, id.(int))
		if err != nil {
			return hcloudutil.ErrorToDiag(err)
		}
		if n == nil {
			return diag.Errorf("no network found with id %d", id)
		}
		setNetworkSchema(d, n)
		return nil
	}
	if name, ok := d.GetOk("name"); ok {
		n, _, err := client.Network.GetByName(ctx, name.(string))
		if err != nil {
			return hcloudutil.ErrorToDiag(err)
		}
		if n == nil {
			return diag.Errorf("no network found with name %s", name)
		}
		setNetworkSchema(d, n)
		return nil
	}

	selector := d.Get("with_selector").(string)
	if selector != "" {
		var allNetworks []*hcloud.Network

		opts := hcloud.NetworkListOpts{
			ListOpts: hcloud.ListOpts{
				LabelSelector: selector,
			},
		}
		allNetworks, err := client.Network.AllWithOpts(ctx, opts)
		if err != nil {
			return hcloudutil.ErrorToDiag(err)
		}
		if len(allNetworks) == 0 {
			return diag.Errorf("no network found for selector %q", selector)
		}
		if len(allNetworks) > 1 {
			return diag.Errorf("more than one network found for selector %q", selector)
		}
		setNetworkSchema(d, allNetworks[0])
		return nil
	}
	return diag.Errorf("please specify an id, a name or a selector to lookup the network")
}

func dataSourceHcloudNetworkListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	selector := d.Get("with_selector").(string)

	opts := hcloud.NetworkListOpts{
		ListOpts: hcloud.ListOpts{
			LabelSelector: selector,
		},
	}
	allNetworks, err := client.Network.AllWithOpts(ctx, opts)
	if err != nil {
		return hcloudutil.ErrorToDiag(err)
	}

	ids := make([]string, len(allNetworks))
	tsNetworks := make([]map[string]interface{}, len(allNetworks))
	for i, firewall := range allNetworks {
		ids[i] = strconv.Itoa(firewall.ID)
		tsNetworks[i] = getNetworkAttributes(firewall)
	}
	d.Set("networks", tsNetworks)
	d.SetId(fmt.Sprintf("%x", sha1.Sum([]byte(strings.Join(ids, "")))))

	return nil
}
