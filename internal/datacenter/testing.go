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

// DatacentersDData defines the fields for the "testdata/d/hcloud_datacenters"
// template.
type DatacentersDData struct {
	testtemplate.DataCommon
}

// TFID returns the data source identifier.
func (d *DatacentersDData) TFID() string {
	return fmt.Sprintf("data.%s.%s", DatacentersDataSourceType, d.RName())
}
