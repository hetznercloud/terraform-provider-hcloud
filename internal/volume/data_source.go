package volume

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
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/datasourceutil"
)

const (
	// DataSourceType is the type name of the Hetzner Cloud Volume data source.
	DataSourceType = "hcloud_volume"

	// DataSourceListType is the type name to receive a list of Hetzner Cloud Volume data source.
	DataSourceListType = "hcloud_volumes"
)

// getCommonDataSchema returns a new common schema used by all volume data sources.
func getCommonDataSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
		},
		"name": {
			Type:     schema.TypeString,
			Computed: true,
			Optional: true,
		},
		"size": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"labels": {
			Type:     schema.TypeMap,
			Computed: true,
		},
		"location": {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
		},
		"server_id": {
			Type:     schema.TypeInt,
			Optional: true,
		},
		"linux_device": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"delete_protection": {
			Type:     schema.TypeBool,
			Computed: true,
		},
	}
}

// DataSource creates a Terraform schema for the hcloud_volume data source.
func DataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudVolumeRead,
		Schema: datasourceutil.MergeSchema(
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

func DataSourceList() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudVolumeListRead,
		Schema: map[string]*schema.Schema{
			"volumes": {
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

func dataSourceHcloudVolumeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	if id, ok := d.GetOk("id"); ok {
		v, _, err := client.Volume.GetByID(ctx, id.(int))
		if err != nil {
			return hcclient.ErrorToDiag(err)
		}
		if v == nil {
			return diag.Errorf("no volume found with id %d", id)
		}
		setVolumeSchema(d, v)
		return nil
	}
	if name, ok := d.GetOk("name"); ok {
		v, _, err := client.Volume.GetByName(ctx, name.(string))
		if err != nil {
			return hcclient.ErrorToDiag(err)
		}
		if v == nil {
			return diag.Errorf("no volume found with name %v", name)
		}
		setVolumeSchema(d, v)
		return nil
	}

	var selector string
	if v := d.Get("with_selector").(string); v != "" {
		selector = v
	} else if v := d.Get("selector").(string); v != "" {
		selector = v
	}
	if selector != "" {
		var allVolumes []*hcloud.Volume

		var statuses []hcloud.VolumeStatus
		for _, status := range d.Get("with_status").([]interface{}) {
			statuses = append(statuses, hcloud.VolumeStatus(status.(string)))
		}

		opts := hcloud.VolumeListOpts{
			ListOpts: hcloud.ListOpts{
				LabelSelector: selector,
			},
			Status: statuses,
		}
		allVolumes, err := client.Volume.AllWithOpts(ctx, opts)
		if err != nil {
			return hcclient.ErrorToDiag(err)
		}
		if len(allVolumes) == 0 {
			return diag.Errorf("no volume found for selector %q", selector)
		}
		if len(allVolumes) > 1 {
			return diag.Errorf("more than one volume found for selector %q", selector)
		}
		setVolumeSchema(d, allVolumes[0])
		return nil
	}
	return diag.Errorf("please specify an id, a name or a selector to lookup the volume")
}

func dataSourceHcloudVolumeListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	selector := d.Get("with_selector").(string)

	statuses := make([]hcloud.VolumeStatus, 0)
	for _, status := range d.Get("with_status").([]interface{}) {
		statuses = append(statuses, hcloud.VolumeStatus(status.(string)))
	}

	opts := hcloud.VolumeListOpts{
		ListOpts: hcloud.ListOpts{
			LabelSelector: selector,
		},
		Status: statuses,
	}
	allVolumes, err := client.Volume.AllWithOpts(ctx, opts)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}

	ids := make([]string, len(allVolumes))
	tfVolume := make([]map[string]interface{}, len(allVolumes))
	for i, volume := range allVolumes {
		ids[i] = strconv.Itoa(volume.ID)
		tfVolume[i] = getVolumeAttributes(volume)
	}
	d.Set("volumes", tfVolume)
	d.SetId(fmt.Sprintf("%x", sha1.Sum([]byte(strings.Join(ids, "")))))

	return nil
}
