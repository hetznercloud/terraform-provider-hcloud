package network_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/e2etests"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/network"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestNetworkSubnetResource_Basic(t *testing.T) {
	var nw hcloud.Network

	resNetwork := &network.RData{
		Name:    "network-test-subnet",
		IPRange: "10.0.0.0/16",
		Labels:  nil,
	}
	resNetwork.SetRName("network-subnet")
	res := &network.RDataSubnet{
		Type:        "cloud",
		NetworkID:   resNetwork.TFID() + ".id",
		NetworkZone: "eu-central",
		IPRange:     "10.0.0.0/24",
	}
	res.SetRName("network-subnet-test")
	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(network.ResourceType, network.ByID(t, &nw)),
		Steps: []resource.TestStep{
			{
				// Create a new Network using the required values
				// only.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_network", resNetwork,
					"testdata/r/hcloud_network_subnet", res,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(resNetwork.TFID(), network.ByID(t, &nw)),
					resource.TestCheckResourceAttr(res.TFID(), "type", res.Type),
					resource.TestCheckResourceAttr(res.TFID(), "ip_range", res.IPRange),
					resource.TestCheckResourceAttr(res.TFID(), "network_zone", res.NetworkZone),
				),
			},
			{
				// Try to import the newly created Network
				ResourceName:      res.TFID(),
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return fmt.Sprintf("%d-%s", nw.ID, res.IPRange), nil
				},
			},
		},
	})
}

func TestNetworkSubnetResource_VSwitch(t *testing.T) {
	var nw hcloud.Network
	vSwitchID := "15074"
	resNetwork := &network.RData{
		Name:    "network-test-vswitch",
		IPRange: "10.0.0.0/16",
		Labels:  nil,
	}
	resNetwork.SetRName("network-subnet-vswitch")
	res := &network.RDataSubnet{
		Type:        "vswitch",
		NetworkID:   resNetwork.TFID() + ".id",
		NetworkZone: "eu-central",
		IPRange:     "10.0.0.0/24",
		VSwitchID:   vSwitchID,
	}
	res.SetRName("network-subnet-vswitch-test")
	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(network.ResourceType, network.ByID(t, &nw)),
		Steps: []resource.TestStep{
			{
				// Create a new Network using the required values
				// only.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_network", resNetwork,
					"testdata/r/hcloud_network_subnet", res,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(resNetwork.TFID(), network.ByID(t, &nw)),
					resource.TestCheckResourceAttr(res.TFID(), "type", res.Type),
					resource.TestCheckResourceAttr(res.TFID(), "ip_range", res.IPRange),
					resource.TestCheckResourceAttr(res.TFID(), "network_zone", res.NetworkZone),
					resource.TestCheckResourceAttr(res.TFID(), "vswitch_id", vSwitchID),
				),
			},
		},
	})
}
