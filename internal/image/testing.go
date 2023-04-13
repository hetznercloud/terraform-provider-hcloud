package image

import (
	"fmt"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

// DData defines the fields for the "testdata/d/hcloud_image"
// template.
type DData struct {
	testtemplate.DataCommon

	ImageID       string
	ImageName     string
	LabelSelector string
	Architecture  hcloud.Architecture
}

// TFID returns the data source identifier.
func (d *DData) TFID() string {
	return fmt.Sprintf("data.%s.%s", DataSourceType, d.RName())
}

// DData defines the fields for the "testdata/d/hcloud_images"
// template.
type DDataList struct {
	testtemplate.DataCommon

	LabelSelector string
	Architecture  hcloud.Architecture
}

// TFID returns the data source identifier.
func (d *DDataList) TFID() string {
	return fmt.Sprintf("data.%s.%s", DataSourceListType, d.RName())
}
