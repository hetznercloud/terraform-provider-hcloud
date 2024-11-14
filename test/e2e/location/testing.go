package location

import (
	"fmt"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/location"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

// DData defines the fields for the "testdata/d/hcloud_location"
// template.
type DData struct {
	testtemplate.DataCommon

	LocationID   string
	LocationName string
}

// TFID returns the data source identifier.
func (d *DData) TFID() string {
	return fmt.Sprintf("data.%s.%s", location.DataSourceType, d.RName())
}

// LocationsDData defines the fields for the "testdata/d/hcloud_locations"
// template.
type DDataList struct {
	testtemplate.DataCommon
}

// TFID returns the data source identifier.
func (d *DDataList) TFID() string {
	return fmt.Sprintf("data.%s.%s", location.DataSourceListType, d.RName())
}
