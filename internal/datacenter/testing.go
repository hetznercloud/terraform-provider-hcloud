package datacenter

import (
	"fmt"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/testtemplate"
)

// DData defines the fields for the "testdata/d/hcloud_datacenter"
// template.
type DData struct {
	testtemplate.DataCommon

	DatacenterID   string
	DatacenterName string
}

// TFID returns the data source identifier.
func (d *DData) TFID() string {
	return fmt.Sprintf("data.%s.%s", DataSourceType, d.RName())
}
