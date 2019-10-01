package hcloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func dataSourceHcloudVolume() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceHcloudVolumeRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Optional: true,
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
			"server": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"linux_device": {
				Type:     schema.TypeString,
				Computed: true,
			},
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
	}
}
func dataSourceHcloudVolumeRead(d *schema.ResourceData, m interface{}) (err error) {
	client := m.(*hcloud.Client)
	ctx := context.Background()
	var v *hcloud.Volume
	if id, ok := d.GetOk("id"); ok {
		v, _, err = client.Volume.GetByID(ctx, id.(int))
		if err != nil {
			return err
		}
		if v == nil {
			return fmt.Errorf("no volume found with id %d", id)
		}
		setVolumeSchema(d, v)
		return
	}
	if name, ok := d.GetOk("name"); ok {
		v, _, err = client.Volume.GetByName(ctx, name.(string))
		if err != nil {
			return err
		}
		if v == nil {
			return fmt.Errorf("no volume found with name %v", name)
		}
		setVolumeSchema(d, v)
		return
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
		allVolumes, err = client.Volume.AllWithOpts(ctx, opts)
		if err != nil {
			return err
		}
		if len(allVolumes) == 0 {
			return fmt.Errorf("no volume found for selector %q", selector)
		}
		if len(allVolumes) > 1 {
			return fmt.Errorf("more than one volume found for selector %q", selector)
		}
		setVolumeSchema(d, allVolumes[0])
		return
	}
	return fmt.Errorf("please specify an id, a name or a selector to lookup the volume")
}
