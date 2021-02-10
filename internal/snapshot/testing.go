package snapshot

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
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

// Sweep removes all Snapshots from the Hetzner Cloud backend.
func Sweep(r string) error {
	client, err := testsupport.CreateClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	images, err := client.Image.AllWithOpts(ctx, hcloud.ImageListOpts{Type: []hcloud.ImageType{hcloud.ImageTypeSnapshot}})
	if err != nil {
		return err
	}

	for _, img := range images {
		if _, err := client.Image.Delete(ctx, img); err != nil {
			return err
		}
	}

	return nil
}

// ByID returns a function that obtains a image by its ID.
func ByID(t *testing.T, img *hcloud.Image) func(*hcloud.Client, int) bool {
	return func(c *hcloud.Client, id int) bool {
		found, _, err := c.Image.GetByID(context.Background(), id)
		if err != nil {
			t.Fatalf("find image %d: %v", id, err)
		}
		if found == nil {
			return false
		}
		if img != nil {
			*img = *found
		}
		return true
	}
}

// RData defines the fields for the "testdata/r/hcloud_snapshot" template.
type RData struct {
	testtemplate.DataCommon

	ServerID    string
	Description string
	Labels      map[string]string
}

// TFID returns the resource identifier.
func (d *RData) TFID() string {
	return fmt.Sprintf("%s.%s", ResourceType, d.RName())
}
