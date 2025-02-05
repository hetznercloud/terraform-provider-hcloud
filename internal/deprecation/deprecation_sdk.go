package deprecation

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

func AddToSchema(s map[string]*schema.Schema) map[string]*schema.Schema {
	s["is_deprecated"] = &schema.Schema{
		Type:     schema.TypeBool,
		Computed: true,
	}
	s["deprecation_announced"] = &schema.Schema{
		Type:     schema.TypeString,
		Computed: true,
		Optional: true,
	}
	s["unavailable_after"] = &schema.Schema{
		Type:     schema.TypeString,
		Computed: true,
		Optional: true,
	}

	return s
}

func SetData(d *schema.ResourceData, r hcloud.Deprecatable) {
	if !r.IsDeprecated() {
		d.Set("is_deprecated", false)
		d.Set("deprecation_announced", nil)
		d.Set("unavailable_after", nil)
	} else {
		d.Set("is_deprecated", true)
		d.Set("deprecation_announced", r.DeprecationAnnounced().Format(time.RFC3339))
		d.Set("unavailable_after", r.UnavailableAfter().Format(time.RFC3339))
	}
}
