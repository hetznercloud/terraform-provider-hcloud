package primaryip

import (
	"context"
	"fmt"
	"testing"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

// ByID returns a function that obtains a primary IP by its ID.
func ByID(t *testing.T, fl *hcloud.PrimaryIP) func(*hcloud.Client, int64) bool {
	return func(c *hcloud.Client, id int64) bool {
		found, _, err := c.PrimaryIP.GetByID(context.Background(), id)
		if err != nil {
			t.Fatalf("find primary ip %d: %v", id, err)
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

// DData defines the fields for the "testdata/d/hcloud_primary_ip"
// template.
type DData struct {
	testtemplate.DataCommon

	PrimaryIPID   string
	PrimaryIPName string
	PrimaryIPIP   string
	LabelSelector string
}

// TFID returns the data source identifier.
func (d *DData) TFID() string {
	return fmt.Sprintf("data.%s.%s", DataSourceType, d.RName())
}

// DDataList defines the fields for the "testdata/d/hcloud_primary_ips" template.
type DDataList struct {
	testtemplate.DataCommon

	LabelSelector string
}

// TFID DDataList the data source identifier.
func (d *DDataList) TFID() string {
	return fmt.Sprintf("data.%s.%s", DataSourceListType, d.RName())
}

// RData defines the fields for the "testdata/r/hcloud_primary_ip" template.
type RData struct {
	testtemplate.DataCommon

	Name             string
	Type             string
	Location         string
	Datacenter       string
	AssigneeType     string
	AssigneeID       string
	Labels           map[string]string
	AutoDelete       bool
	DeleteProtection bool
}

// TFID returns the resource identifier.
func (d *RData) TFID() string {
	return fmt.Sprintf("%s.%s", ResourceType, d.RName())
}
