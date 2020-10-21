package server_test

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/network"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/server"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
	"github.com/stretchr/testify/assert"
)

func TestAccHcloudServerNetwork_NetworkID(t *testing.T) {
	var (
		nw hcloud.Network
		s  hcloud.Server
	)

	netRes := &network.RData{
		Name:    "test-network",
		IPRange: "10.0.0.0/16",
	}
	netRes.SetRName("test-network")
	subNetRes := &network.RDataSubnet{
		Type:        "cloud",
		NetworkID:   netRes.TFID() + ".id",
		NetworkZone: "eu-central",
		IPRange:     "10.0.1.0/24",
	}
	subNetRes.SetRName("test-network-subnet")
	sRes := &server.RData{
		Name:         "s-network-test",
		Type:         "cx11",
		LocationName: "nbg1",
		Image:        "ubuntu-20.04",
	}
	sRes.SetRName("s-network-test")
	sNRes := &server.RDataNetwork{
		Name:      "test-network",
		ServerID:  sRes.TFID() + ".id",
		NetworkID: netRes.TFID() + ".id",
		IP:        "10.0.1.5",
		DependsOn: []string{subNetRes.TFID()},
	}
	tmplMan := testtemplate.Manager{}
	resource.Test(t, resource.TestCase{
		PreCheck:     testsupport.AccTestPreCheck(t),
		Providers:    testsupport.AccTestProviders(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_network", netRes,
					"testdata/r/hcloud_network_subnet", subNetRes,
					"testdata/r/hcloud_server", sRes,
					"testdata/r/hcloud_server_network", sNRes,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(netRes.TFID(), network.ByID(t, &nw)),
					testsupport.CheckResourceExists(sRes.TFID(), server.ByID(t, &s)),
					testsupport.LiftTCF(hasServerNetwork(t, &s, &nw)),
					resource.TestCheckResourceAttr(
						server.NetworkResourceType+".test-network", "ip", "10.0.1.5"),
				),
			},
			{
				// Try to import the newly created Server
				ResourceName:      server.NetworkResourceType + ".test-network",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return fmt.Sprintf("%d-%d", s.ID, nw.ID), nil
				},
			},
		},
	})
}

func TestAccHcloudServerNetwork_SubNetID(t *testing.T) {
	var (
		nw hcloud.Network
		s  hcloud.Server
	)

	netRes := &network.RData{
		Name:    "test-network",
		IPRange: "10.0.0.0/16",
	}
	netRes.SetRName("test-network")
	subNetRes := &network.RDataSubnet{
		Type:        "cloud",
		NetworkID:   netRes.TFID() + ".id",
		NetworkZone: "eu-central",
		IPRange:     "10.0.1.0/24",
	}
	subNetRes.SetRName("test-network-subnet")
	sRes := &server.RData{
		Name:         "s-network-test",
		Type:         "cx11",
		LocationName: "nbg1",
		Image:        "ubuntu-20.04",
	}
	sRes.SetRName("s-network-test")

	tmplMan := testtemplate.Manager{}
	resource.Test(t, resource.TestCase{
		PreCheck:     testsupport.AccTestPreCheck(t),
		Providers:    testsupport.AccTestProviders(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_network", netRes,
					"testdata/r/hcloud_network_subnet", subNetRes,
					"testdata/r/hcloud_server", sRes,
					"testdata/r/hcloud_server_network", &server.RDataNetwork{
						Name:     "test-network",
						ServerID: sRes.TFID() + ".id",
						SubNetID: subNetRes.TFID() + ".id",
						IP:       "10.0.1.5",
					},
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(netRes.TFID(), network.ByID(t, &nw)),
					testsupport.CheckResourceExists(sRes.TFID(), server.ByID(t, &s)),
					testsupport.LiftTCF(hasServerNetwork(t, &s, &nw)),
				),
			},
		},
	})
}

func hasServerNetwork(t *testing.T, s *hcloud.Server, nw *hcloud.Network) func() error {
	return func() error {
		var privNet *hcloud.ServerPrivateNet
		for _, n := range s.PrivateNet {
			if n.Network.ID == nw.ID {
				privNet = &n
				break
			}
		}
		if privNet == nil {
			return fmt.Errorf("server has no private network")
		}
		assert.Equal(t, "10.0.1.5", privNet.IP.String())
		return nil
	}
}
