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
	ServerTypesDataSourceType = "hcloud_server_types"
)

// DataSourcecreates a new Terraform schema for the hcloud_server_type data
// source.
func DataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudServerTypeRead,
		Schema: map[string]*schema.Schema{
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

func setServerTypeSchema(data *schema.ResourceData, d *hcloud.ServerType) {
	data.SetId(strconv.Itoa(d.ID))
	data.Set("name", d.Name)
	data.Set("description", d.Description)
	data.Set("cores", d.Cores)
	data.Set("memory", d.Memory)
	data.Set("disk", d.Disk)
	data.Set("description", d.Description)
	data.Set("storage_type", d.StorageType)
	data.Set("cpu_type", d.CPUType)
}

// ServerTypesDataSource creates a new Terraform schema for the
// hcloud_server_types data source.
func ServerTypesDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudServerTypesRead,
		Schema: map[string]*schema.Schema{
			"server_type_ids": {
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

func dataSourceHcloudServerTypesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)
	sts, err := client.ServerType.All(ctx)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}

	names := make([]string, len(sts))
	descriptions := make([]string, len(sts))
	ids := make([]string, len(sts))
	for i, v := range sts {
		ids[i] = strconv.Itoa(v.ID)
		descriptions[i] = v.Description
		names[i] = v.Name
	}

	d.SetId(fmt.Sprintf("%x", sha1.Sum([]byte(strings.Join(ids, "")))))
	d.Set("server_type_ids", ids)
	d.Set("names", names)
	d.Set("descriptions", descriptions)
	return nil
}
