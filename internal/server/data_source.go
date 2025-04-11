package server

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/datasourceutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/merge"
)

const (
	// DataSourceType is the type name of the Hetzner Cloud server resource.
	DataSourceType = "hcloud_server"

	// DataSourceListType is the type name to receive a list of Hetzner Cloud server resources.
	DataSourceListType = "hcloud_servers"
)

// getCommonDataSchema returns a new common schema used by all server data sources.
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
		"server_type": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"image": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"location": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"datacenter": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"backup_window": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"backups": {
			Type:     schema.TypeBool,
			Computed: true,
		},
		"ipv4_address": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"ipv6_address": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"ipv6_network": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"status": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"iso": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"rescue": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"labels": {
			Type:     schema.TypeMap,
			Computed: true,
		},
		"firewall_ids": {
			Type:     schema.TypeList,
			Computed: true,
			Elem:     &schema.Schema{Type: schema.TypeInt},
		},
		"network": {
			Type:     schema.TypeSet,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"network_id": {
						Type:     schema.TypeInt,
						Computed: true,
					},
					"ip": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"alias_ips": {
						Type:     schema.TypeSet,
						Elem:     &schema.Schema{Type: schema.TypeString},
						Computed: true,
					},
					"mac_address": {
						Type:     schema.TypeString,
						Computed: true,
					},
				},
			},
		},
		"placement_group_id": {
			Type:     schema.TypeInt,
			Optional: true,
		},
		"delete_protection": {
			Type:     schema.TypeBool,
			Computed: true,
		},
		"rebuild_protection": {
			Type:     schema.TypeBool,
			Computed: true,
		},
		"primary_disk_size": {
			Type:     schema.TypeInt,
			Computed: true,
		},
	}
}

// DataSource creates a new Terraform schema for the hcloud_server resource.
func DataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudServerRead,
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
				"with_status": {
					Type: schema.TypeList,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Optional: true,
				},
			},
		),
	}
}

// DataSourceList creates a new Terraform schema for the hcloud_servers resource.
func DataSourceList() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudServerListRead,
		Schema: map[string]*schema.Schema{
			"servers": {
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
			"with_status": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
		},
	}
}

func dataSourceServerItem() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudServerListRead,
		Schema:      getCommonDataSchema(),
	}
}

func dataSourceHcloudServerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	if id, ok := d.GetOk("id"); ok {
		s, _, err := client.Server.GetByID(ctx, util.CastInt64(id))
		if err != nil {
			return hcloudutil.ErrorToDiag(err)
		}
		if s == nil {
			return diag.Errorf("no Server found with id %d", id)
		}
		setServerSchema(d, s, true)
		return nil
	}

	if name, ok := d.GetOk("name"); ok {
		s, _, err := client.Server.GetByName(ctx, name.(string))
		if err != nil {
			return hcloudutil.ErrorToDiag(err)
		}
		if s == nil {
			return diag.Errorf("no Server found with name %s", name)
		}
		setServerSchema(d, s, true)
		return nil
	}

	var selector string
	if v := d.Get("with_selector").(string); v != "" {
		selector = v
	} else if v := d.Get("selector").(string); v != "" {
		selector = v
	}
	if selector != "" {
		var allServers []*hcloud.Server
		var statuses []hcloud.ServerStatus
		for _, status := range d.Get("with_status").([]interface{}) {
			statuses = append(statuses, hcloud.ServerStatus(status.(string)))
		}

		opts := hcloud.ServerListOpts{
			ListOpts: hcloud.ListOpts{
				LabelSelector: selector,
			},
			Status: statuses,
		}
		allServers, err := client.Server.AllWithOpts(ctx, opts)
		if err != nil {
			return hcloudutil.ErrorToDiag(err)
		}
		if len(allServers) == 0 {
			return diag.Errorf("no Server found for selector %q", selector)
		}
		if len(allServers) > 1 {
			return diag.Errorf("more than one Server found for selector %q", selector)
		}
		setServerSchema(d, allServers[0], true)
		return nil
	}

	return diag.Errorf("please specify a id, name or a selector to lookup the Server")
}

func dataSourceHcloudServerListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	selector := d.Get("with_selector").(string)

	statuses := make([]hcloud.ServerStatus, 0)
	for _, status := range d.Get("with_status").([]interface{}) {
		statuses = append(statuses, hcloud.ServerStatus(status.(string)))
	}

	opts := hcloud.ServerListOpts{
		ListOpts: hcloud.ListOpts{
			LabelSelector: selector,
		},
		Status: statuses,
	}
	allServers, err := client.Server.AllWithOpts(ctx, opts)
	if err != nil {
		return hcloudutil.ErrorToDiag(err)
	}

	ids := make([]string, len(allServers))
	tfServers := make([]map[string]interface{}, len(allServers))
	for i, server := range allServers {
		ids[i] = util.FormatID(server.ID)
		tfServers[i] = getServerAttributes(d, server, true)
	}
	d.Set("servers", tfServers)
	d.SetId(datasourceutil.ListID(ids))

	return nil
}
