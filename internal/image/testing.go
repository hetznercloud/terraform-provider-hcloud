package image

import (
	"fmt"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

// DData defines the fields for the "testdata/d/hcloud_image"
// template.
type DData struct {
	testtemplate.DataCommon

	ID                string
	Name              string
	WithSelector      string
	WithArchitecture  hcloud.Architecture
	IncludeDeprecated bool
	MostRecent        *bool

	Raw string
}

// TFID returns the data source identifier.
func (d *DData) TFID() string {
	return fmt.Sprintf("data.%s.%s", DataSourceType, d.RName())
}

// DDataList defines the fields for the "testdata/d/hcloud_images"
// template.
type DDataList struct {
	testtemplate.DataCommon

	WithSelector      string
	WithArchitecture  hcloud.Architecture
	WithStatus        hcloud.ImageStatus
	IncludeDeprecated bool

	Raw string
}

// TFID returns the data source identifier.
func (d *DDataList) TFID() string {
	return fmt.Sprintf("data.%s.%s", DataSourceListType, d.RName())
}
