package iso

import (
	"context"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/datasourceutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/merge"
)

const (
	// DataSourceType is the type name of the Hetzner Cloud ISO resource.
	DataSourceType = "hcloud_iso"

	// DataSourceListType is the type name to receive a list of Hetzner Cloud ISO resources.
	DataSourceListType = "hcloud_isos"
)

// getCommonDataSchema returns a new common schema used by all iso data sources.
func getCommonDataSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
		},
		"type": {
			Type:     schema.TypeString,
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
		"architecture": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"deprecated_announced": {
			Type:     schema.TypeString,
			Computed: true,
			Optional: true,
		},
		"unavailable_after": {
			Type:     schema.TypeString,
			Computed: true,
			Optional: true,
		},
	}
}

// DataSource creates a Terraform schema for the hcloud_iso data source.
func DataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudISORead,
		Schema: merge.Maps(
			getCommonDataSchema(),
			map[string]*schema.Schema{
				"name_prefix": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"with_architecture": {
					Type:     schema.TypeString,
					Default:  hcloud.ArchitectureX86,
					Optional: true,
				},
				"include_architecture_wildcard": {
					Type:     schema.TypeBool,
					Default:  false,
					Optional: true,
				},
			},
		),
	}
}

func DataSourceList() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudISOListRead,
		Schema: map[string]*schema.Schema{
			"isos": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: getCommonDataSchema(),
				},
			},
			"name_prefix": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"with_architecture": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
			"include_architecture_wildcard": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
		},
	}
}

func dataSourceHcloudISORead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)
	if id, ok := d.GetOk("id"); ok {
		i, _, err := client.ISO.GetByID(ctx, util.CastInt64(id))
		if err != nil {
			return hcloudutil.ErrorToDiag(err)
		}
		if i == nil {
			return diag.Errorf("no ISO found with id %d", id)
		}
		setISOSchema(d, i)
		return nil
	}

	opts := hcloud.ISOListOpts{}

	name := d.Get("name").(string)
	if name != "" {
		opts.Name = name
	}

	namePrefix, useNamePrefix := d.Get("name_prefix").(string)

	// Resources can be selected either by name
	if name == "" || (!useNamePrefix || namePrefix == "") {
		diag.Errorf("please specify an id, a name or a name prefix to lookup the ISO")
	}

	var isoType hcloud.ISOType
	if isoTypeStr := d.Get("type").(string); isoTypeStr != "" {
		switch isoTypeStr {
		case string(hcloud.ISOTypePublic), string(hcloud.ISOTypePrivate):
			isoType = hcloud.ISOType(isoTypeStr)
		default:
			return diag.Errorf("unknown ISO types %s", isoTypeStr)
		}
	}

	log.Printf("Arches: %+v", d.Get("with_architecture"))

	architecture := hcloud.Architecture(d.Get("with_architecture").(string))
	if architecture != "" {
		opts.Architecture = []hcloud.Architecture{architecture}
	}

	if d.Get("include_architecture_wildcard").(bool) {
		opts.IncludeArchitectureWildcard = true
	}

	allISOs, err := client.ISO.AllWithOpts(ctx, opts)
	if err != nil {
		return hcloudutil.ErrorToDiag(err)
	}
	if len(allISOs) == 0 {
		return diag.Errorf("no ISO found matching the selection")
	}
	if len(allISOs) > 1 {
		if useNamePrefix {
			// If we have a name prefix, we filter the list to only include the first
			// ISO that matches the prefix.
			allISOs = filterByNamePrefix(allISOs, namePrefix)
		}
		if isoType != "" {
			allISOs = filterByType(allISOs, isoType)
		}
		if len(allISOs) == 0 {
			return diag.Errorf("no ISO found matching the selection with name prefix %s", name)
		} else if len(allISOs) > 1 {
			return diag.Errorf("multiple ISOs found matching the selection with name prefix %s, please use a more specific prefix", name)
		}
		log.Printf("[INFO] %d ISOs found, using %d as the more accurate one", len(allISOs), allISOs[0].ID)
	}
	setISOSchema(d, allISOs[0])
	return nil
}

func dataSourceHcloudISOListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	architectures := make([]hcloud.Architecture, 0)
	for _, arch := range d.Get("with_architecture").(*schema.Set).List() {
		architectures = append(architectures, hcloud.Architecture(arch.(string)))
	}

	var isoType hcloud.ISOType
	if isoTypeStr := d.Get("type").(string); isoTypeStr != "" {
		switch isoTypeStr {
		case string(hcloud.ISOTypePublic), string(hcloud.ISOTypePrivate):
			isoType = hcloud.ISOType(isoTypeStr)
		default:
			return diag.Errorf("unknown ISO types %s", isoTypeStr)
		}
	}

	opts := hcloud.ISOListOpts{
		Architecture: architectures,
	}
	if d.Get("include_architecture_wildcard").(bool) {
		opts.IncludeArchitectureWildcard = true
	}

	allISOs, err := client.ISO.AllWithOpts(ctx, opts)
	if err != nil {
		return hcloudutil.ErrorToDiag(err)
	}

	namePrefix := d.Get("name_prefix").(string)
	if namePrefix != "" {
		allISOs = filterByNamePrefix(allISOs, namePrefix)
	}
	if isoType != "" {
		allISOs = filterByType(allISOs, isoType)
	}

	ids := make([]string, len(allISOs))
	tfISOs := make([]map[string]interface{}, len(allISOs))
	for i, iso := range allISOs {
		ids[i] = util.FormatID(iso.ID)
		tfISOs[i] = getISOAttributes(iso)
	}
	d.Set("isos", tfISOs)
	d.SetId(datasourceutil.ListID(ids))

	return nil
}

func filterByNamePrefix(isoList []*hcloud.ISO, namePrefix string) []*hcloud.ISO {
	slices.SortFunc(isoList, func(a, b *hcloud.ISO) int {
		return strings.Compare(a.Name, b.Name)
	})
	seq := func(yield func(*hcloud.ISO) bool) {
		for _, i := range isoList {
			if !strings.HasPrefix(i.Name, namePrefix) {
				continue
			}
			// If the yield function returns false, we stop iterating.
			if !yield(i) {
				return
			}
		}
	}
	return slices.Collect(seq)
}

func filterByType(isoList []*hcloud.ISO, isoType hcloud.ISOType) []*hcloud.ISO {
	seq := func(yield func(*hcloud.ISO) bool) {
		for _, i := range isoList {
			if i.Type != isoType {
				continue
			}
			// If the yield function returns false, we stop iterating.
			if !yield(i) {
				return
			}
		}
	}
	return slices.Collect(seq)
}

func setISOSchema(d *schema.ResourceData, i *hcloud.ISO) {
	util.SetSchemaFromAttributes(d, getISOAttributes(i))
}

func getISOAttributes(i *hcloud.ISO) map[string]interface{} {
	res := map[string]interface{}{
		"id":          i.ID,
		"type":        i.Type,
		"name":        i.Name,
		"description": i.Description,
	}
	if i.Architecture != nil {
		res["architecture"] = i.Architecture
	}
	if !i.IsDeprecated() {
		res["deprecated_announced"] = i.DeprecationAnnounced().Format(time.RFC3339)
		res["unavailable_after"] = i.UnavailableAfter().Format(time.RFC3339)
	}
	return res
}
