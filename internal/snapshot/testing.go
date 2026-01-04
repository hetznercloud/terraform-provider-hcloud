package snapshot

import (
	"context"
	"fmt"
	"testing"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

// ByID returns a function that obtains a image by its ID.
func ByID(t *testing.T, img *hcloud.Image) func(*hcloud.Client, int64) bool {
	return func(c *hcloud.Client, id int64) bool {
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
