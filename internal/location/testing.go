package location

import (
	"fmt"

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
	return fmt.Sprintf("data.%s.%s", DataSourceType, d.RName())
}

// LocationsDData defines the fields for the "testdata/d/hcloud_locations"
// template.
type LocationsDData struct {
	testtemplate.DataCommon
}

// TFID returns the data source identifier.
func (d *LocationsDData) TFID() string {
	return fmt.Sprintf("data.%s.%s", LocationsDataSourceType, d.RName())
}
