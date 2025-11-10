package storagebox

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func init() {
	resource.AddTestSweepers(ResourceType, &resource.Sweeper{
		Name:         ResourceType,
		Dependencies: []string{},
		F:            Sweep,
	})
}

// Sweep removes all Storage Boxes from the Hetzner API.
func Sweep(r string) error {
	client, err := testsupport.CreateClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	storageBoxes, err := client.StorageBox.All(ctx)
	if err != nil {
		return err
	}

	actions := make([]*hcloud.Action, 0, len(storageBoxes))
	for _, o := range storageBoxes {
		result, _, err := client.StorageBox.Delete(ctx, o)
		if err != nil {
			return err
		}
		actions = append(actions, result.Action)
	}

	if err := client.Action.WaitFor(ctx, actions...); err != nil {
		return err
	}

	return nil
}

// GetAPIResource returns a [testsupport.GetAPIResourceFunc] for [hcloud.StorageBox].
func GetAPIResource() testsupport.GetAPIResourceFunc[hcloud.StorageBox] {
	return func(c *hcloud.Client, attrs map[string]string) (*hcloud.StorageBox, error) {
		result, _, err := c.StorageBox.Get(context.Background(), attrs["id"])
		return result, err
	}
}

// DData defines the fields for the "testdata/d/hcloud_storage_box" template.
type DData struct {
	testtemplate.DataCommon

	ID            string
	Name          string
	LabelSelector string
}

// TFID returns the data source identifier.
func (d *DData) TFID() string {
	return fmt.Sprintf("data.%s.%s", DataSourceType, d.RName())
}

// DDataList defines the fields for the "testdata/d/hcloud_storage_boxes" template.
type DDataList struct {
	testtemplate.DataCommon

	LabelSelector string
}

// TFID returns the data source identifier.
func (d *DDataList) TFID() string {
	return fmt.Sprintf("data.%s.%s", DataSourceListType, d.RName())
}

// RData defines the fields for the "testdata/r/hcloud_storage_box" template.
type RData struct {
	testtemplate.DataCommon
	schema.StorageBox
	Password string
	SSHKeys  []string
	Raw      string
}

// TFID returns the resource identifier.
func (d *RData) TFID() string {
	return fmt.Sprintf("%s.%s", ResourceType, d.RName())
}
