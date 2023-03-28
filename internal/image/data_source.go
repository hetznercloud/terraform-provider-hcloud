package image

import (
	"context"
	"crypto/sha1"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/hcclient"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/datasourceutil"
)

const (
	// DataSourceType is the type name of the Hetzner Cloud image resource.
	DataSourceType = "hcloud_image"

	// DataSourceListType is the type name to receive a list of Hetzner Cloud image resources.
	DataSourceListType = "hcloud_images"
)

// getCommonDataSchema returns a new common schema used by all image data sources.
func getCommonDataSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
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
		"architecture": {
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
	}
}

// DataSource creates a Terraform schema for the hcloud_image data source.
func DataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudImageRead,
		Schema: datasourceutil.MergeSchema(
			getCommonDataSchema(),
			map[string]*schema.Schema{
				"most_recent": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"with_selector": {
					Type:          schema.TypeString,
					Optional:      true,
					ConflictsWith: []string{"selector", "name"},
				},
				"with_status": {
					Type: schema.TypeList,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Optional: true,
				},
				"with_architecture": {
					Type:     schema.TypeString,
					Default:  hcloud.ArchitectureX86,
					Optional: true,
				},
			},
		),
	}
}

func DataSourceList() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudImageListRead,
		Schema: map[string]*schema.Schema{
			"images": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: getCommonDataSchema(),
				},
			},
			"most_recent": {
				Type:     schema.TypeBool,
				Optional: true,
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
			"with_architecture": {
				Type: schema.TypeSet,
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

	opts := hcloud.ImageListOpts{}

	name := d.Get("name").(string)
	if name != "" {
		opts.Name = name
	}

	var selector string
	if v := d.Get("with_selector").(string); v != "" {
		selector = v
	} else if v := d.Get("selector").(string); v != "" {
		selector = v
	}
	if selector != "" {
		opts.LabelSelector = selector
	}

	// Resources can be selected either by name or selector
	if name != "" && selector != "" {
		diag.Errorf("you can only use one of name and with_selector")
	}
	if name == "" && selector == "" {
		diag.Errorf("please specify an id, a name or a selector to lookup the image")
	}

	statuses := make([]hcloud.ImageStatus, 0)
	for _, status := range d.Get("with_status").([]interface{}) {
		statuses = append(statuses, hcloud.ImageStatus(status.(string)))
	}
	opts.Status = statuses

	log.Printf("Arches: %+v", d.Get("with_architecture"))

	architecture := hcloud.Architecture(d.Get("with_architecture").(string))
	if architecture != "" {
		opts.Architecture = []hcloud.Architecture{architecture}
	}

	allImages, err := client.Image.AllWithOpts(ctx, opts)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	if len(allImages) == 0 {
		return diag.Errorf("no image found matching the selection")
	}
	if len(allImages) > 1 {
		if _, ok := d.GetOk("most_recent"); !ok {
			return diag.Errorf("more than one image found")
		}
		sortImageListByCreated(allImages)
		log.Printf("[INFO] %d images found, using %d as the most recent one", len(allImages), allImages[0].ID)
	}
	setImageSchema(d, allImages[0])
	return nil
}

func dataSourceHcloudImageListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	selector := d.Get("with_selector").(string)

	statuses := make([]hcloud.ImageStatus, 0)
	for _, status := range d.Get("with_status").([]interface{}) {
		statuses = append(statuses, hcloud.ImageStatus(status.(string)))
	}

	architectures := make([]hcloud.Architecture, 0)
	for _, arch := range d.Get("with_architecture").(*schema.Set).List() {
		architectures = append(architectures, hcloud.Architecture(arch.(string)))
	}

	opts := hcloud.ImageListOpts{
		ListOpts:     hcloud.ListOpts{LabelSelector: selector},
		Status:       statuses,
		Architecture: architectures,
	}
	allImages, err := client.Image.AllWithOpts(ctx, opts)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}

	if _, ok := d.GetOk("most_recent"); ok {
		sortImageListByCreated(allImages)
	}

	ids := make([]string, len(allImages))
	tfImages := make([]map[string]interface{}, len(allImages))
	for i, image := range allImages {
		ids[i] = strconv.Itoa(image.ID)
		tfImages[i] = getImageAttributes(image)
	}
	d.Set("images", tfImages)
	d.SetId(fmt.Sprintf("%x", sha1.Sum([]byte(strings.Join(ids, "")))))

	return nil
}

func sortImageListByCreated(imageList []*hcloud.Image) {
	sort.Slice(imageList, func(i, j int) bool {
		return imageList[i].Created.After(imageList[j].Created)
	})
}

func setImageSchema(d *schema.ResourceData, i *hcloud.Image) {
	for key, val := range getImageAttributes(i) {
		if key == "id" {
			d.SetId(strconv.Itoa(val.(int)))
		} else {
			d.Set(key, val)
		}
	}
}

func getImageAttributes(i *hcloud.Image) map[string]interface{} {
	res := map[string]interface{}{
		"id":           i.ID,
		"type":         i.Type,
		"name":         i.Name,
		"created":      i.Created.Format(time.RFC3339),
		"description":  i.Description,
		"os_flavor":    i.OSFlavor,
		"os_version":   i.OSVersion,
		"architecture": i.Architecture,
		"rapid_deploy": i.RapidDeploy,
		"labels":       i.Labels,
	}

	if !i.Deprecated.IsZero() {
		res["deprecated"] = i.Deprecated.Format(time.RFC3339)
	}

	return res
}
