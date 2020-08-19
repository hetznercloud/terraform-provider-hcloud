package network_test

import (
	"github.com/terraform-providers/terraform-provider-hcloud/internal/network"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/testsupport"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/testtemplate"
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
	resource.Test(t, resource.TestCase{
		PreCheck:     testsupport.AccTestPreCheck(t),
		Providers:    testsupport.AccTestProviders(),
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
		},
	})
}
