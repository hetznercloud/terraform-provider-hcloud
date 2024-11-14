package network

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/network"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func init() {
	resource.AddTestSweepers(network.ResourceType, &resource.Sweeper{
		Name: network.ResourceType,
		F:    Sweep,
	})
}

// Sweep removes all Networks from the Hetzner Cloud backend.
func Sweep(r string) error {
	client, err := testsupport.CreateClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	networks, err := client.Network.All(ctx)
	if err != nil {
		return err
	}

	for _, nw := range networks {
		if _, err := client.Network.Delete(ctx, nw); err != nil {
			return err
		}
	}

	return nil
}

// ByID returns a function that obtains a network by its ID.
func ByID(t *testing.T, nw *hcloud.Network) func(*hcloud.Client, int) bool {
	return func(c *hcloud.Client, id int) bool {
		found, _, err := c.Network.GetByID(context.Background(), id)
		if err != nil {
			t.Fatalf("network by ID: %d: %v", id, err)
		}
		if found == nil {
			return false
		}
		if nw != nil {
			*nw = *found
		}
		return true
	}
}

// DData defines the fields for the "testdata/d/hcloud_network"
// template.
type DData struct {
	testtemplate.DataCommon

	NetworkID     string
	NetworkName   string
	LabelSelector string
}

// TFID returns the data source identifier.
func (d *DData) TFID() string {
	return fmt.Sprintf("data.%s.%s", network.DataSourceType, d.RName())
}

// RData defines the fields for the "testdata/r/hcloud_network" template.
type RData struct {
	testtemplate.DataCommon

	Name                  string
	IPRange               string
	Labels                map[string]string
	DeleteProtection      bool
	ExposeRoutesToVSwitch bool
}

// TFID returns the resource identifier.
func (d *RData) TFID() string {
	return fmt.Sprintf("%s.%s", network.ResourceType, d.RName())
}

// DDataList defines the fields for the "testdata/d/hcloud_networks" template.
type DDataList struct {
	testtemplate.DataCommon

	LabelSelector string
}

// TFID DDataList the data source identifier.
func (d *DDataList) TFID() string {
	return fmt.Sprintf("data.%s.%s", network.DataSourceListType, d.RName())
}

// RDataSubnet defines the fields for the "testdata/r/hcloud_network_subnet"
// template.
type RDataSubnet struct {
	testtemplate.DataCommon

	Type        string
	NetworkID   string
	NetworkZone string
	IPRange     string
	VSwitchID   string
}

// TFID returns the resource identifier.
func (d *RDataSubnet) TFID() string {
	return fmt.Sprintf("%s.%s", network.SubnetResourceType, d.RName())
}

// RDataRoute defines the fields for the "testdata/r/hcloud_network_route"
// template.
type RDataRoute struct {
	testtemplate.DataCommon

	NetworkID   string
	Destination string
	Gateway     string
}

// TFID returns the resource identifier.
func (d *RDataRoute) TFID() string {
	return fmt.Sprintf("%s.%s", network.RouteResourceType, d.RName())
}
