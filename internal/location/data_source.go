package location

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
	// DataSourceType is the type name of the Hetzner Cloud location datasource.
	DataSourceType = "hcloud_location"

	// DataSourceListType is the type name of the Hetzner Cloud location datasource.
	DataSourceListType = "hcloud_locations"
)

// getCommonDataSchema returns a new common schema used by all location data sources.
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
		"country": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"city": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"latitude": {
			Type:     schema.TypeFloat,
			Computed: true,
		},
		"longitude": {
			Type:     schema.TypeFloat,
			Computed: true,
		},
		"network_zone": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

// DataSource creates a new Terraform schema for the hcloud_location
// datasource.
func DataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudLocationRead,
		Schema:      getCommonDataSchema(),
	}
}

// DataSourceList creates a new Terraform schema for the hcloud_locations
// datasource.
func DataSourceList() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudLocationListRead,
		Schema: map[string]*schema.Schema{
			"location_ids": {
				Type:       schema.TypeList,
				Optional:   true,
				Deprecated: "Use locations list instead",
				Elem:       &schema.Schema{Type: schema.TypeString},
			},
			"names": {
				Type:       schema.TypeList,
				Computed:   true,
				Deprecated: "Use locations list instead",
				Elem:       &schema.Schema{Type: schema.TypeString},
			},
			"descriptions": {
				Type:       schema.TypeList,
				Computed:   true,
				Deprecated: "Use locations list instead",
				Elem:       &schema.Schema{Type: schema.TypeString},
			},
			"locations": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: getCommonDataSchema(),
				},
			},
		},
	}
}

func dataSourceHcloudLocationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	if id, ok := d.GetOk("id"); ok {
		l, _, err := client.Location.GetByID(ctx, id.(int))
		if err != nil {
			return hcclient.ErrorToDiag(err)
		}
		if l == nil {
			return diag.Errorf("no location found with id %d", id)
		}
		setLocationSchema(d, l)
		return nil
	}
	if name, ok := d.GetOk("name"); ok {
		l, _, err := client.Location.GetByName(ctx, name.(string))
		if err != nil {
			return hcclient.ErrorToDiag(err)
		}
		if l == nil {
			return diag.Errorf("no location found with name %v", name)
		}
		setLocationSchema(d, l)
		return nil
	}

	return diag.Errorf("please specify an id, or a name to lookup for a location")
}

func setLocationSchema(d *schema.ResourceData, l *hcloud.Location) {
	for key, val := range getLocationAttributes(l) {
		if key == "id" {
			d.SetId(strconv.Itoa(val.(int)))
		} else {
			d.Set(key, val)
		}
	}
}

func getLocationAttributes(l *hcloud.Location) map[string]interface{} {
	return map[string]interface{}{
		"id":           l.ID,
		"name":         l.Name,
		"description":  l.Description,
		"country":      l.Country,
		"city":         l.City,
		"latitude":     l.Latitude,
		"longitude":    l.Longitude,
		"network_zone": l.NetworkZone,
	}
}

func dataSourceHcloudLocationListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	allLocations, err := client.Location.All(ctx)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}

	names := make([]string, len(allLocations))
	descriptions := make([]string, len(allLocations))
	ids := make([]string, len(allLocations))
	tfLocations := make([]map[string]interface{}, len(allLocations))
	for i, location := range allLocations {
		ids[i] = strconv.Itoa(location.ID)
		descriptions[i] = location.Description
		names[i] = location.Name

		tfLocations[i] = getLocationAttributes(location)
	}

	d.SetId(fmt.Sprintf("%x", sha1.Sum([]byte(strings.Join(ids, "")))))
	d.Set("location_ids", ids)
	d.Set("names", names)
	d.Set("descriptions", descriptions)
	d.Set("locations", tfLocations)

	return nil
}
