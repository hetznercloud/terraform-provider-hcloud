package hcloud

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func dataSourceHcloudImage() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceHcloudImageRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"type": {
				Type:     schema.TypeString,
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
			"created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"os_flavor": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"os_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"rapid_deploy": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"deprecated": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceHcloudImageRead(d *schema.ResourceData, m interface{}) (err error) {
	client := m.(*hcloud.Client)
	ctx := context.Background()
	var i *hcloud.Image
	if id, ok := d.GetOk("id"); ok {
		i, _, err = client.Image.GetByID(ctx, id.(int))
		if err != nil {
			return err
		}
		if i == nil {
			return fmt.Errorf("no image found with id %d", id)
		}
		setImageSchema(d, i)
		return
	}
	if name, ok := d.GetOk("name"); ok {
		i, _, err = client.Image.GetByName(ctx, name.(string))
		if err != nil {
			return err
		}
		if i == nil {
			return fmt.Errorf("no image found with name %v", name)
		}
		setImageSchema(d, i)
		return
	}
	return fmt.Errorf("please specify an id or name to lookup the image")
}

func setImageSchema(d *schema.ResourceData, i *hcloud.Image) {
	d.SetId(strconv.Itoa(i.ID))
	d.Set("type", i.Type)
	d.Set("name", i.Name)
	d.Set("created", i.Created)
	d.Set("description", i.Description)
	d.Set("os_flavor", i.OSFlavor)
	d.Set("os_version", i.OSVersion)
	d.Set("rapid_deploy", i.RapidDeploy)
	d.Set("deprecated", i.Deprecated)
}
