package iso

import (
	"fmt"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

// DData defines the fields for the "testdata/d/hcloud_iso"
// template.
type DData struct {
	testtemplate.DataCommon

	IsoID                       string
	IsoName                     string
	IsoNamePrefix               string
	IsoType                     hcloud.ISOType
	Architecture                hcloud.Architecture
	IncludeArchitectureWildcard bool
}

// TFID returns the data source identifier.
func (d *DData) TFID() string {
	return fmt.Sprintf("data.%s.%s", DataSourceType, d.RName())
}

// DDataList defines the fields for the "testdata/d/hcloud_isos"
// template.
type DDataList struct {
	testtemplate.DataCommon

	IsoNamePrefix               string
	IsoType                     hcloud.ISOType
	Architecture                hcloud.Architecture
	IncludeArchitectureWildcard bool
}

// TFID returns the data source identifier.
func (d *DDataList) TFID() string {
	return fmt.Sprintf("data.%s.%s", DataSourceListType, d.RName())
}
