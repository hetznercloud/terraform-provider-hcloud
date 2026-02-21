package storageboxsubaccount

import (
	"context"
	"fmt"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
)

// GetAPIResource returns a [testsupport.GetAPIResourceFunc] for [hcloud.StorageBoxSubaccount].
func GetAPIResource() testsupport.GetAPIResourceFunc[hcloud.StorageBoxSubaccount] {
	return func(c *hcloud.Client, attrs map[string]string) (*hcloud.StorageBoxSubaccount, error) {
		storageBoxID, err := util.ParseID(attrs["storage_box_id"])
		if err != nil {
			return nil, err
		}
		subaccountID, err := util.ParseID(attrs["id"])
		if err != nil {
			return nil, err
		}

		result, _, err := c.StorageBox.GetSubaccountByID(context.Background(), &hcloud.StorageBox{ID: storageBoxID}, subaccountID)
		return result, err
	}
}

// DData defines the fields for the "testdata/d/hcloud_storage_box_subaccount" template.
type DData struct {
	testtemplate.DataCommon

	StorageBox    string
	ID            string
	Username      string
	LabelSelector string
}

// TFID returns the data source identifier.
func (d *DData) TFID() string {
	return fmt.Sprintf("data.%s.%s", DataSourceType, d.RName())
}

// DDataList defines the fields for the "testdata/d/hcloud_storage_box_subaccounts" template.
type DDataList struct {
	testtemplate.DataCommon

	StorageBox    string
	LabelSelector string
}

// TFID returns the data source identifier.
func (d *DDataList) TFID() string {
	return fmt.Sprintf("data.%s.%s", DataSourceListType, d.RName())
}

// RData defines the fields for the "testdata/r/hcloud_storage_box_subaccount" template.
type RData struct {
	testtemplate.DataCommon

	StorageBox    string
	HomeDirectory string
	Name          string
	Password      string // nolint: gosec
	Description   string
	Labels        map[string]string

	Raw string
}

// TFID returns the resource identifier.
func (d *RData) TFID() string {
	return fmt.Sprintf("%s.%s", ResourceType, d.RName())
}
