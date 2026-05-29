package primaryip

import (
	"context"
	"fmt"
	"testing"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

// GetAPIResource returns a [testsupport.GetAPIResourceFunc] for [hcloud.PrimaryIP].
func GetAPIResource() testsupport.GetAPIResourceFunc[hcloud.PrimaryIP] {
	return func(c *hcloud.Client, attrs map[string]string) (*hcloud.PrimaryIP, error) {
		result, _, err := c.PrimaryIP.Get(context.Background(), attrs["id"])
		return result, err
	}
}

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

	Raw string
}

// TFID returns the data source identifier.
func (d *DData) TFID() string {
	return fmt.Sprintf("data.%s.%s", DataSourceType, d.RName())
}

// DDataList defines the fields for the "testdata/d/hcloud_primary_ips" template.
type DDataList struct {
	testtemplate.DataCommon

	LabelSelector string

	Raw string
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
	AutoDelete       *bool
	DeleteProtection bool

	Raw string
}

// TFID returns the resource identifier.
func (d *RData) TFID() string {
	return fmt.Sprintf("%s.%s", ResourceType, d.RName())
}

type Blueprint struct {
	PrimaryIPv4A *RData
	PrimaryIPv4B *RData
	PrimaryIPv6C *RData
	PrimaryIPv6D *RData
}

func NewBlueprint(t *testing.T) *Blueprint {
	t.Helper()

	b := &Blueprint{}

	b.PrimaryIPv4A = &RData{
		Name:     "a-ipv4",
		Type:     "ipv4",
		Location: teste2e.TestLocationName,
	}
	b.PrimaryIPv4A.SetRName("primary_ipv4_a")

	b.PrimaryIPv4B = &RData{
		Name:     "b-ipv4",
		Type:     "ipv4",
		Location: teste2e.TestLocationName,
	}
	b.PrimaryIPv4B.SetRName("primary_ipv4_b")

	b.PrimaryIPv6C = &RData{
		Name:     "c-ipv6",
		Type:     "ipv6",
		Location: teste2e.TestLocationName,
	}
	b.PrimaryIPv6C.SetRName("primary_ipv6_c")

	b.PrimaryIPv6D = &RData{
		Name:     "d-ipv6",
		Type:     "ipv6",
		Location: teste2e.TestLocationName,
	}
	b.PrimaryIPv6D.SetRName("primary_ipv6_d")

	return b
}
