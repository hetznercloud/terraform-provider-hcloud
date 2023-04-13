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

func TestNetworkRouteResource_Basic(t *testing.T) {
	var nw hcloud.Network

	resNetwork := &network.RData{
		Name:    "network-test-route",
		IPRange: "10.0.0.0/16",
	}
	resNetwork.SetRName("network-route")
	res := &network.RDataRoute{
		NetworkID:   resNetwork.TFID() + ".id",
		Destination: "10.100.1.0/24",
		Gateway:     "10.0.1.1",
	}
	res.SetRName("network-route-test")
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
					"testdata/r/hcloud_network_route", res,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(resNetwork.TFID(), network.ByID(t, &nw)),
					resource.TestCheckResourceAttr(res.TFID(), "destination", res.Destination),
					resource.TestCheckResourceAttr(res.TFID(), "gateway", res.Gateway),
				),
			},
			{
				// Try to import the newly created Network
				ResourceName:      res.TFID(),
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return fmt.Sprintf("%d-%s", nw.ID, res.Destination), nil
				},
			},
		},
	})
}
