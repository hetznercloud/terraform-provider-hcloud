package network

import (
	"context"
	"fmt"
	"testing"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

// ByID returns a function that obtains a network by its ID.
func ByID(t *testing.T, nw *hcloud.Network) func(*hcloud.Client, int64) bool {
	return func(c *hcloud.Client, id int64) bool {
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
	return fmt.Sprintf("data.%s.%s", DataSourceType, d.RName())
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
	return fmt.Sprintf("%s.%s", ResourceType, d.RName())
}

// DDataList defines the fields for the "testdata/d/hcloud_networks" template.
type DDataList struct {
	testtemplate.DataCommon

	LabelSelector string
}

// TFID DDataList the data source identifier.
func (d *DDataList) TFID() string {
	return fmt.Sprintf("data.%s.%s", DataSourceListType, d.RName())
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

	DependsOn []string
}

// TFID returns the resource identifier.
func (d *RDataSubnet) TFID() string {
	return fmt.Sprintf("%s.%s", SubnetResourceType, d.RName())
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
	return fmt.Sprintf("%s.%s", RouteResourceType, d.RName())
}

type Blueprint struct {
	NetworkA *RData
	SubnetA1 *RDataSubnet
	SubnetA2 *RDataSubnet

	NetworkB *RData
	SubnetB1 *RDataSubnet
	SubnetB2 *RDataSubnet
}

func NewBlueprint(t *testing.T) *Blueprint {
	t.Helper()

	b := &Blueprint{}

	// Network A
	b.NetworkA = &RData{
		Name:    "a",
		IPRange: "10.0.0.0/16",
	}
	b.NetworkA.SetRName("network_a")

	b.SubnetA1 = &RDataSubnet{
		NetworkID:   b.NetworkA.TFID() + ".id",
		NetworkZone: "eu-central",
		IPRange:     "10.0.1.0/24",
		Type:        "cloud",
	}
	b.SubnetA1.SetRName("subnet_a1")

	b.SubnetA2 = &RDataSubnet{
		NetworkID:   b.NetworkA.TFID() + ".id",
		NetworkZone: "eu-central",
		IPRange:     "10.0.2.0/24",
		Type:        "cloud",
	}
	b.SubnetA2.SetRName("subnet_a2")

	// Network B
	b.NetworkB = &RData{
		Name:    "b",
		IPRange: "172.16.0.0/16",
	}
	b.NetworkB.SetRName("network_b")

	b.SubnetB1 = &RDataSubnet{
		NetworkID:   b.NetworkB.TFID() + ".id",
		NetworkZone: "eu-central",
		IPRange:     "172.16.1.0/24",
		Type:        "cloud",
	}
	b.SubnetB1.SetRName("subnet_b1")

	b.SubnetB2 = &RDataSubnet{
		NetworkID:   b.NetworkB.TFID() + ".id",
		NetworkZone: "eu-central",
		IPRange:     "172.16.2.0/24",
		Type:        "cloud",
	}
	b.SubnetB2.SetRName("subnet_b2")

	return b
}
