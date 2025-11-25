package storageboxsnapshot

import (
	"context"
	"fmt"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
)

// GetAPIResource returns a [testsupport.GetAPIResourceFunc] for [hcloud.StorageBoxSnapshot].
func GetAPIResource() testsupport.GetAPIResourceFunc[hcloud.StorageBoxSnapshot] {
	return func(c *hcloud.Client, attrs map[string]string) (*hcloud.StorageBoxSnapshot, error) {
		storageBoxID, err := util.ParseID(attrs["storage_box_id"])
		if err != nil {
			return nil, err
		}
		snapshotID, err := util.ParseID(attrs["id"])
		if err != nil {
			return nil, err
		}

		result, _, err := c.StorageBox.GetSnapshotByID(context.Background(), &hcloud.StorageBox{ID: storageBoxID}, snapshotID)
		return result, err
	}
}

// DData defines the fields for the "testdata/d/hcloud_storage_box_snapshot" template.
type DData struct {
	testtemplate.DataCommon

	StorageBox    string
	ID            string
	Name          string
	LabelSelector string
}

// TFID returns the data source identifier.
func (d *DData) TFID() string {
	return fmt.Sprintf("data.%s.%s", DataSourceType, d.RName())
}

// DDataList defines the fields for the "testdata/d/hcloud_storage_box_snapshots" template.
type DDataList struct {
	testtemplate.DataCommon

	StorageBox    string
	LabelSelector string
}

// TFID returns the data source identifier.
func (d *DDataList) TFID() string {
	return fmt.Sprintf("data.%s.%s", DataSourceListType, d.RName())
}

// RData defines the fields for the "testdata/r/hcloud_storage_box_snapshot" template.
type RData struct {
	testtemplate.DataCommon

	StorageBox  string
	Description string
	Labels      map[string]string

	Raw string
}

// TFID returns the resource identifier.
func (d *RData) TFID() string {
	return fmt.Sprintf("%s.%s", ResourceType, d.RName())
}
