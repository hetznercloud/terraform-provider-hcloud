package network_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/network"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccNetworkSubnetResource_Basic(t *testing.T) {
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
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(network.ResourceType, network.ByID(t, &nw)),
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
				ImportStateIdFunc: func(_ *terraform.State) (string, error) {
					return fmt.Sprintf("%d-%s", nw.ID, res.IPRange), nil
				},
			},
		},
	})
}

func TestAccNetworkSubnetResource_VSwitch(t *testing.T) {
	t.Skip("No VSwitch available in test account")

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
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(network.ResourceType, network.ByID(t, &nw)),
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
