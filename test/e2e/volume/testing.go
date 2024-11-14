package volume

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
	volume2 "github.com/hetznercloud/terraform-provider-hcloud/internal/volume"
)

func init() {
	resource.AddTestSweepers(volume2.ResourceType, &resource.Sweeper{
		Name:         volume2.ResourceType,
		Dependencies: []string{},
		F:            Sweep,
	})
}

// Sweep removes all Volumes from the Hetzner Cloud backend.
func Sweep(r string) error {
	client, err := testsupport.CreateClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	servers, err := client.Volume.All(ctx)
	if err != nil {
		return err
	}

	for _, srv := range servers {
		if _, err := client.Volume.Delete(ctx, srv); err != nil {
			return err
		}
	}

	return nil
}

// ByID returns a function that obtains a Volume by its ID.
func ByID(t *testing.T, fl *hcloud.Volume) func(*hcloud.Client, int) bool {
	return func(c *hcloud.Client, id int) bool {
		found, _, err := c.Volume.GetByID(context.Background(), id)
		if err != nil {
			t.Fatalf("find floating ip %d: %v", id, err)
		}
		if found == nil {
			return false
		}
		if fl != nil {
			*fl = *found
		}
		return true
	}
}

// DData defines the fields for the "testdata/d/hcloud_volume"
// template.
type DData struct {
	testtemplate.DataCommon

	VolumeID      string
	VolumeName    string
	LabelSelector string
}

// TFID returns the data source identifier.
func (d *DData) TFID() string {
	return fmt.Sprintf("data.%s.%s", volume2.DataSourceType, d.RName())
}

// DData defines the fields for the "testdata/d/hcloud_volumes"
// template.
type DDataList struct {
	testtemplate.DataCommon

	LabelSelector string
}

// TFID returns the data source identifier.
func (d *DDataList) TFID() string {
	return fmt.Sprintf("data.%s.%s", volume2.DataSourceListType, d.RName())
}

// RData defines the fields for the "testdata/r/hcloud_volume" template.
type RData struct {
	testtemplate.DataCommon

	Name             string
	Size             int
	LocationName     string
	Labels           map[string]string
	ServerID         string
	DeleteProtection bool
}

// TFID returns the resource identifier.
func (d *RData) TFID() string {
	return fmt.Sprintf("%s.%s", volume2.ResourceType, d.RName())
}

// RDataAttachment defines the fields for the "testdata/r/hcloud_volume_attachment" template.
type RDataAttachment struct {
	testtemplate.DataCommon

	VolumeID string
	ServerID string
}

// TFID returns the resource identifier.
func (d *RDataAttachment) TFID() string {
	return fmt.Sprintf("%s.%s", volume2.AttachmentResourceType, d.RName())
}

// BasicRData is a resource for use in volume related test.
func BasicRData() *RData {
	return &RData{
		Name:         "basic-volume",
		LocationName: teste2e.TestLocationName,
		Size:         10,
	}
}
