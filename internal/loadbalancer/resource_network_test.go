package loadbalancer_test

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/loadbalancer"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/network"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
	"github.com/stretchr/testify/assert"
)

func TestAccHcloudLoadBalancerNetwork_NetworkID(t *testing.T) {
	var (
		nw hcloud.Network
		lb hcloud.LoadBalancer
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
	lbRes := &loadbalancer.RData{
		Name:        "lb-network-test",
		Type:        "lb11",
		NetworkZone: "eu-central",
	}

	tmplMan := testtemplate.Manager{}
	resource.Test(t, resource.TestCase{
		PreCheck:     testsupport.AccTestPreCheck(t),
		Providers:    testsupport.AccTestProviders(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_network", netRes,
					"testdata/r/hcloud_network_subnet", subNetRes,
					"testdata/r/hcloud_load_balancer", lbRes,
					"testdata/r/hcloud_load_balancer_network", &loadbalancer.RDataNetwork{
						Name:                  "test-network",
						LoadBalancerID:        lbRes.TFID() + ".id",
						NetworkID:             netRes.TFID() + ".id",
						IP:                    "10.0.1.5",
						EnablePublicInterface: false,
						DependsOn:             []string{subNetRes.TFID()},
					},
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(netRes.TFID(), network.ByID(t, &nw)),
					testsupport.CheckResourceExists(lbRes.TFID(), loadbalancer.ByID(t, &lb)),
					testsupport.LiftTCF(hasLoadBalancerNetwork(t, &lb, &nw)),
					resource.TestCheckResourceAttr(
						loadbalancer.NetworkResourceType+".test-network", "ip", "10.0.1.5"),
					resource.TestCheckResourceAttr(
						loadbalancer.NetworkResourceType+".test-network", "enable_public_interface", "false"),
				),
			},
			{
				// Try to import the newly created Server
				ResourceName:      loadbalancer.NetworkResourceType + ".test-network",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return fmt.Sprintf("%d-%d", lb.ID, nw.ID), nil
				},
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_network", netRes,
					"testdata/r/hcloud_network_subnet", subNetRes,
					"testdata/r/hcloud_load_balancer", lbRes,
					"testdata/r/hcloud_load_balancer_network", &loadbalancer.RDataNetwork{
						Name:                  "test-network",
						LoadBalancerID:        "hcloud_load_balancer.lb-network-test.id",
						NetworkID:             "hcloud_network.test-network.id",
						IP:                    "10.0.1.5",
						EnablePublicInterface: true,
						DependsOn:             []string{subNetRes.TFID()},
					},
				),
				Check: resource.TestCheckResourceAttr(
					loadbalancer.NetworkResourceType+".test-network", "enable_public_interface", "true"),
			},
		},
	})
}

func TestAccHcloudLoadBalancerNetwork_SubNetID(t *testing.T) {
	var (
		nw hcloud.Network
		lb hcloud.LoadBalancer
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
	lbRes := &loadbalancer.RData{
		Name:        "lb-network-test",
		Type:        "lb11",
		NetworkZone: "eu-central",
	}

	tmplMan := testtemplate.Manager{}
	resource.Test(t, resource.TestCase{
		PreCheck:     testsupport.AccTestPreCheck(t),
		Providers:    testsupport.AccTestProviders(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_network", netRes,
					"testdata/r/hcloud_network_subnet", subNetRes,
					"testdata/r/hcloud_load_balancer", lbRes,
					"testdata/r/hcloud_load_balancer_network", &loadbalancer.RDataNetwork{
						Name:           "test-network",
						LoadBalancerID: lbRes.TFID() + ".id",
						SubNetID:       subNetRes.TFID() + ".id",
						IP:             "10.0.1.5",
					},
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(netRes.TFID(), network.ByID(t, &nw)),
					testsupport.CheckResourceExists(lbRes.TFID(), loadbalancer.ByID(t, &lb)),
					testsupport.LiftTCF(hasLoadBalancerNetwork(t, &lb, &nw)),
				),
			},
		},
	})
}

func hasLoadBalancerNetwork(t *testing.T, lb *hcloud.LoadBalancer, nw *hcloud.Network) func() error {
	return func() error {
		var privNet *hcloud.LoadBalancerPrivateNet
		for _, n := range lb.PrivateNet {
			if n.Network.ID == nw.ID {
				privNet = &n
				break
			}
		}
		if privNet == nil {
			return fmt.Errorf("load balancer has no private network")
		}
		assert.Equal(t, "10.0.1.5", privNet.IP.String())
		return nil
	}
}
