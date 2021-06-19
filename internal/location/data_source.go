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

	// LocationsDataSourceType is the type name of the Hetzner Cloud location datasource.
	LocationsDataSourceType = "hcloud_locations"
)

// DataSource creates a new Terraform schema for the hcloud_location
// datasource.
func DataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudLocationRead,
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
	d.SetId(strconv.Itoa(l.ID))
	d.Set("name", l.Name)
	d.Set("description", l.Description)
	d.Set("country", l.Country)
	d.Set("city", l.City)
	d.Set("latitude", l.Latitude)
	d.Set("longitude", l.Longitude)
}

// LocationsDataSource creates a new Terraform schema for the hcloud_locations
// datasource.
func LocationsDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudLocationsRead,
		Schema: map[string]*schema.Schema{
			"location_ids": {
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

func dataSourceHcloudLocationsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)
	ls, err := client.Location.All(ctx)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}

	names := make([]string, len(ls))
	descriptions := make([]string, len(ls))
	ids := make([]string, len(ls))
	for i, v := range ls {
		ids[i] = strconv.Itoa(v.ID)
		descriptions[i] = v.Description
		names[i] = v.Name
	}

	d.SetId(fmt.Sprintf("%x", sha1.Sum([]byte(strings.Join(ids, "")))))
	d.Set("location_ids", ids)
	d.Set("names", names)
	d.Set("descriptions", descriptions)
	return nil
}
