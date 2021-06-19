package image

import (
	"context"
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/hcclient"
)

// DataSourceType is the type name of the Hetzner Cloud image resource.
const DataSourceType = "hcloud_image"

// DataSource creates a Terraform schema for the hcloud_image data source.
func DataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudImageRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
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
			"labels": {
				Type:     schema.TypeMap,
				Computed: true,
			},
			"selector": {
				Type:          schema.TypeString,
				Optional:      true,
				Deprecated:    "Please use the with_selector property instead.",
				ConflictsWith: []string{"with_selector"},
			},
			"most_recent": {
				Type:     schema.TypeBool,
				Optional: true,
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

func dataSourceHcloudImageRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)
	if id, ok := d.GetOk("id"); ok {
		i, _, err := client.Image.GetByID(ctx, id.(int))
		if err != nil {
			return hcclient.ErrorToDiag(err)
		}
		if i == nil {
			return diag.Errorf("no image found with id %d", id)
		}
		setImageSchema(d, i)
		return nil
	}
	if name, ok := d.GetOk("name"); ok {
		i, _, err := client.Image.GetByName(ctx, name.(string))
		if err != nil {
			return hcclient.ErrorToDiag(err)
		}
		if i == nil {
			return diag.Errorf("no image found with name %v", name)
		}
		setImageSchema(d, i)
		return nil
	}
	var selector string
	if v := d.Get("with_selector").(string); v != "" {
		selector = v
	} else if v := d.Get("selector").(string); v != "" {
		selector = v
	}
	if selector != "" {
		var allImages []*hcloud.Image

		var statuses []hcloud.ImageStatus
		for _, status := range d.Get("with_status").([]interface{}) {
			statuses = append(statuses, hcloud.ImageStatus(status.(string)))
		}

		opts := hcloud.ImageListOpts{ListOpts: hcloud.ListOpts{LabelSelector: selector}, Status: statuses}
		allImages, err := client.Image.AllWithOpts(ctx, opts)
		if err != nil {
			return hcclient.ErrorToDiag(err)
		}
		if len(allImages) == 0 {
			return diag.Errorf("no image found for selector %q", selector)
		}
		if len(allImages) > 1 {
			if _, ok := d.GetOk("most_recent"); !ok {
				return diag.Errorf("more than one image found for selector %q", selector)
			}
			sortImageListByCreated(allImages)
			log.Printf("[INFO] %d images found for selector %q, using %d as the most recent one", len(allImages), selector, allImages[0].ID)
		}
		setImageSchema(d, allImages[0])
		return nil
	}
	return diag.Errorf("please specify an id, a name or a selector to lookup the image")
}

func sortImageListByCreated(imageList []*hcloud.Image) {
	sort.Slice(imageList, func(i, j int) bool {
		return imageList[i].Created.After(imageList[j].Created)
	})
}

func setImageSchema(d *schema.ResourceData, i *hcloud.Image) {
	d.SetId(strconv.Itoa(i.ID))
	d.Set("type", i.Type)
	d.Set("name", i.Name)
	d.Set("created", i.Created.Format(time.RFC3339))
	d.Set("description", i.Description)
	d.Set("os_flavor", i.OSFlavor)
	d.Set("os_version", i.OSVersion)
	d.Set("rapid_deploy", i.RapidDeploy)
	if !i.Deprecated.IsZero() {
		d.Set("deprecated", i.Deprecated.Format(time.RFC3339))
	}
	d.Set("labels", i.Labels)
}
