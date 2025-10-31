package zonerecord

import (
	"context"
	"fmt"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

// GetAPIResource returns a [testsupport.GetAPIResourceFunc] for [hcloud.ZoneRRSet].
func GetAPIResource() testsupport.GetAPIResourceFunc[hcloud.ZoneRRSet] {
	return func(c *hcloud.Client, attrs map[string]string) (*hcloud.ZoneRRSet, error) {
		result, _, err := c.Zone.GetRRSetByID(context.Background(), &hcloud.Zone{Name: attrs["zone"]}, attrs["id"])
		return result, err
	}
}

// RData defines the fields for the "testdata/r/hcloud_zone_record" template.
type RData struct {
	testtemplate.DataCommon
	Raw string

	Zone string
	schema.ZoneRRSet
}

// TFID returns the resource identifier.
func (d *RData) TFID() string {
	return fmt.Sprintf("%s.%s", ResourceType, d.RName())
}
