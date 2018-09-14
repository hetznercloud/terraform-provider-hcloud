package hcloud

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

// TODO
func resourceVolume() *schema.Resource {
	return &schema.Resource{
		Create: notImplementedYet,
		Read:   notImplementedYet,
		Update: notImplementedYet,
		Delete: notImplementedYet,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"size": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"location": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"server_id": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
			},
		},
	}
}
func notImplementedYet(d *schema.ResourceData, m interface{}) error {
	return fmt.Errorf("This function is currently not implemented.") // TODO

}
func setVolumeSchema(d *schema.ResourceData, s *hcloud.Volume) {
	d.SetId(strconv.Itoa(s.ID))
	d.Set("name", s.Name)
	d.Set("size", s.Size)
	d.Set("labels", s.Labels)
}
