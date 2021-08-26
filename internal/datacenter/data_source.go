package datacenter

import (
	"context"
	"crypto/sha1"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/hcclient"
)

const (
	// DataSourceType is the type name of the Hetzner Cloud Datacenter datasource.
	DataSourceType = "hcloud_datacenter"

	// DataSourceListType is the type name of the Hetzner Cloud Datacenters datasource.
	DataSourceListType = "hcloud_datacenters"
)

// getCommonDataSchema returns a new common schema used by all datacenter data sources.
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
		"location": {
			Type:     schema.TypeMap,
			Computed: true,
		},
		"supported_server_type_ids": {
			Type:     schema.TypeList,
			Computed: true,
			Elem:     &schema.Schema{Type: schema.TypeInt},
		},
		"available_server_type_ids": {
			Type:     schema.TypeList,
			Computed: true,
			Elem:     &schema.Schema{Type: schema.TypeInt},
		},
	}
}

// DataSource creates a new Terraform schema for the Hetzner Cloud Datacenter
// data source.
func DataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudDatacenterRead,
		Schema:      getCommonDataSchema(),
	}
}

// DataSourceList creates a new Terraform schema for the Hetzner Cloud Datacenters data source.
func DataSourceList() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudDatacenterListRead,
		Schema: map[string]*schema.Schema{
			"datacenter_ids": {
				Type:       schema.TypeList,
				Optional:   true,
				Deprecated: "Use datacenters list instead",
				Elem:       &schema.Schema{Type: schema.TypeString},
			},
			"names": {
				Type:       schema.TypeList,
				Computed:   true,
				Deprecated: "Use datacenters list instead",
				Elem:       &schema.Schema{Type: schema.TypeString},
			},
			"descriptions": {
				Type:       schema.TypeList,
				Computed:   true,
				Deprecated: "Use datacenters list instead",
				Elem:       &schema.Schema{Type: schema.TypeString},
			},
			"datacenters": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: getCommonDataSchema(),
				},
			},
		},
	}
}

func dataSourceHcloudDatacenterRead(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	if id, ok := data.GetOk("id"); ok {
		d, _, err := client.Datacenter.GetByID(ctx, id.(int))
		if err != nil {
			return hcclient.ErrorToDiag(err)
		}
		if d == nil {
			return diag.Errorf("no datacenter found with id %d", id)
		}
		setDatacenterSchema(data, d)
		return nil
	}
	if name, ok := data.GetOk("name"); ok {
		d, _, err := client.Datacenter.GetByName(ctx, name.(string))
		if err != nil {
			return hcclient.ErrorToDiag(err)
		}
		if d == nil {
			return diag.Errorf("no datacenter found with name %v", name)
		}
		setDatacenterSchema(data, d)
		return nil
	}

	return diag.Errorf("please specify an id, or a name to lookup for a datacenter")
}

func setDatacenterSchema(d *schema.ResourceData, dc *hcloud.Datacenter) {
	for key, val := range getDatacenterAttributes(dc) {
		if key == "id" {
			d.SetId(strconv.Itoa(val.(int)))
		} else {
			d.Set(key, val)
		}
	}
}

func getDatacenterAttributes(dc *hcloud.Datacenter) map[string]interface{} {
	supported := make([]int, len(dc.ServerTypes.Supported))

	for i, v := range dc.ServerTypes.Supported {
		supported[i] = v.ID
	}
	available := make([]int, len(dc.ServerTypes.Available))
	for i, v := range dc.ServerTypes.Available {
		available[i] = v.ID
	}
	sort.Ints(available)
	sort.Ints(supported)

	return map[string]interface{}{
		"id":          dc.ID,
		"name":        dc.Name,
		"description": dc.Description,
		"location": map[string]string{
			"id":          strconv.Itoa(dc.Location.ID),
			"name":        dc.Location.Name,
			"description": dc.Location.Description,
			"country":     dc.Location.Country,
			"city":        dc.Location.City,
			"latitude":    fmt.Sprintf("%f", dc.Location.Latitude),
			"longitude":   fmt.Sprintf("%f", dc.Location.Longitude),
		},
		"supported_server_type_ids": supported,
		"available_server_type_ids": available,
	}
}

func dataSourceHcloudDatacenterListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)
	dcs, err := client.Datacenter.All(ctx)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}

	names := make([]string, len(dcs))
	descriptions := make([]string, len(dcs))
	ids := make([]string, len(dcs))
	tfDatacenters := make([]map[string]interface{}, len(dcs))
	for i, datacenter := range dcs {
		ids[i] = strconv.Itoa(datacenter.ID)
		descriptions[i] = datacenter.Description
		names[i] = datacenter.Name

		tfDatacenters[i] = getDatacenterAttributes(datacenter)
	}
	d.SetId(fmt.Sprintf("%x", sha1.Sum([]byte(strings.Join(ids, "")))))
	d.Set("datacenter_ids", ids)
	d.Set("names", names)
	d.Set("descriptions", descriptions)
	d.Set("datacenters", tfDatacenters)

	return nil
}
