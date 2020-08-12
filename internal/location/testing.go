package location

import (
	"fmt"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/testtemplate"
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
