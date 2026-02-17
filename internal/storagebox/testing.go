package storagebox

import (
	"context"
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

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
	Password string // nolint: gosec
	SSHKeys  []string
	Raw      string
}

// TFID returns the resource identifier.
func (d *RData) TFID() string {
	return fmt.Sprintf("%s.%s", ResourceType, d.RName())
}

func GeneratePassword(t *testing.T) string {
	t.Helper()

	characterSets := [4]string{
		"abcdefghijklmnopqrstuvwxyz",
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ",
		"01234567890",
		"!$%/()=?+#-.",
	}

	password := ""

	for _, chars := range characterSets {
		for i := 0; i < 32; i++ {
			password += string(chars[rand.IntN(len(chars))]) // nolint:gosec // Only used for tests
		}
	}

	return password
}
