package zonerrset

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

// DData defines the fields for the "testdata/d/hcloud_zone_rrset" template.
type DData struct {
	testtemplate.DataCommon

	Zone          string
	ID            string
	Name          string
	Type          string
	LabelSelector string
}

// TFID returns the data source identifier.
func (d *DData) TFID() string {
	return fmt.Sprintf("data.%s.%s", DataSourceType, d.RName())
}

// DDataList defines the fields for the "testdata/d/hcloud_zone_rrsets"
// template.
type DDataList struct {
	testtemplate.DataCommon

	Zone          string
	LabelSelector string
}

// TFID returns the data source identifier.
func (d *DDataList) TFID() string {
	return fmt.Sprintf("data.%s.%s", DataSourceListType, d.RName())
}

// RData defines the fields for the "testdata/r/hcloud_zone_rrset" template.
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
