package datacenter

import (
	"fmt"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
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

// DDataList defines the fields for the "testdata/d/hcloud_datacenters" template.
type DDataList struct {
	testtemplate.DataCommon
}

// TFID returns the data source identifier.
func (d *DDataList) TFID() string {
	return fmt.Sprintf("data.%s.%s", DataSourceListType, d.RName())
}
