package servertype

import (
	"context"
	"crypto/sha1"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/hcclient"
)

const (
	// DataSourceType is the type name of the Hetzner Cloud Server Type
	// data source.
	DataSourceType = "hcloud_server_type"

	// ServerTypesDataSourceType is the type name of the Hetzner Cloud Server Types
	// data source.
	DataSourceListType = "hcloud_server_types"
)

// getCommonDataSchema returns a new common schema used by all server type data sources.
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
		"description": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"cores": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"memory": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"disk": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"storage_type": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"cpu_type": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"architecture": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

// DataSource creates a new Terraform schema for the hcloud_server_type data
// source.
func DataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudServerTypeRead,
		Schema:      getCommonDataSchema(),
	}
}

// ServerTypesDataSource creates a new Terraform schema for the
// hcloud_server_types data source.
func ServerTypesDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudServerTypeListRead,
		Schema: map[string]*schema.Schema{
			"server_types": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: getCommonDataSchema(),
				},
			},
			"server_type_ids": {
				Type:       schema.TypeList,
				Optional:   true,
				Deprecated: "Use server_types list instead",
				Elem:       &schema.Schema{Type: schema.TypeString},
			},
			"names": {
				Type:       schema.TypeList,
				Computed:   true,
				Deprecated: "Use server_types list instead",
				Elem:       &schema.Schema{Type: schema.TypeString},
			},
			"descriptions": {
				Type:       schema.TypeList,
				Computed:   true,
				Deprecated: "Use server_types list instead",
				Elem:       &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceHcloudServerTypeRead(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	if id, ok := data.GetOk("id"); ok {
		d, _, err := client.ServerType.GetByID(ctx, id.(int))
		if err != nil {
			return hcclient.ErrorToDiag(err)
		}
		if d == nil {
			return diag.Errorf("no server type found with id %d", id)
		}
		setServerTypeSchema(data, d)
		return nil
	}
	if name, ok := data.GetOk("name"); ok {
		d, _, err := client.ServerType.GetByName(ctx, name.(string))
		if err != nil {
			return hcclient.ErrorToDiag(err)
		}
		if d == nil {
			return diag.Errorf("no server type found with name %v", name)
		}
		setServerTypeSchema(data, d)
		return nil
	}

	return diag.Errorf("please specify an id, or a name to lookup for a server type")
}

func setServerTypeSchema(d *schema.ResourceData, t *hcloud.ServerType) {
	for key, val := range getServerTypeAttributes(t) {
		if key == "id" {
			d.SetId(strconv.Itoa(val.(int)))
		} else {
			d.Set(key, val)
		}
	}
}

func getServerTypeAttributes(t *hcloud.ServerType) map[string]interface{} {
	return map[string]interface{}{
		"id":           t.ID,
		"name":         t.Name,
		"description":  t.Description,
		"cores":        t.Cores,
		"memory":       t.Memory,
		"disk":         t.Disk,
		"storage_type": t.StorageType,
		"cpu_type":     t.CPUType,
		"architecture": t.Architecture,
	}
}

func dataSourceHcloudServerTypeListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)
	allServerTypes, err := client.ServerType.All(ctx)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}

	names := make([]string, len(allServerTypes))
	descriptions := make([]string, len(allServerTypes))
	ids := make([]string, len(allServerTypes))
	tfServerTypes := make([]map[string]interface{}, len(allServerTypes))
	for i, serverType := range allServerTypes {
		ids[i] = strconv.Itoa(serverType.ID)
		descriptions[i] = serverType.Description
		names[i] = serverType.Name

		tfServerTypes[i] = getServerTypeAttributes(serverType)
	}

	d.SetId(fmt.Sprintf("%x", sha1.Sum([]byte(strings.Join(ids, "")))))
	d.Set("server_type_ids", ids)
	d.Set("names", names)
	d.Set("descriptions", descriptions)
	d.Set("server_types", tfServerTypes)

	return nil
}
