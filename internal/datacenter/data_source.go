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
	// DataSourceType is the type name of the Hetzner Cloud Datacenter
	// datasource.
	DataSourceType = "hcloud_datacenter"

	// DatacentersDataSourceType is the type name of the Hetzner Cloud
	// Datacenters datasource.
	DatacentersDataSourceType = "hcloud_datacenters"
)

// DataSource creates a new Terraform schema for the Hetzner Cloud Datacenter
// data source.
func DataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudDatacenterRead,
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
	d.SetId(strconv.Itoa(dc.ID))
	d.Set("name", dc.Name)
	d.Set("description", dc.Description)
	d.Set("location", map[string]string{
		"id":          strconv.Itoa(dc.Location.ID),
		"name":        dc.Location.Name,
		"description": dc.Location.Description,
		"country":     dc.Location.Country,
		"city":        dc.Location.City,
		"latitude":    fmt.Sprintf("%f", dc.Location.Latitude),
		"longitude":   fmt.Sprintf("%f", dc.Location.Longitude),
	})
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
	d.Set("supported_server_type_ids", supported)
	d.Set("available_server_type_ids", available)
}

// DatacentersDataSource creates a new Terraform schema for the Hetzner Cloud
// Datacenters data source.
func DatacentersDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudDatacentersRead,
		Schema: map[string]*schema.Schema{
			"datacenter_ids": {
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

func dataSourceHcloudDatacentersRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)
	dcs, err := client.Datacenter.All(ctx)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}

	names := make([]string, 0, len(dcs))
	descriptions := make([]string, 0, len(dcs))
	ids := make([]string, 0, len(dcs))
	for _, v := range dcs {
		ids = append(ids, strconv.Itoa(v.ID))
		descriptions = append(descriptions, v.Description)
		names = append(names, v.Name)
	}
	d.SetId(fmt.Sprintf("%x", sha1.Sum([]byte(strings.Join(ids, "")))))
	d.Set("datacenter_ids", ids)
	d.Set("names", names)
	d.Set("descriptions", descriptions)
	return nil
}
