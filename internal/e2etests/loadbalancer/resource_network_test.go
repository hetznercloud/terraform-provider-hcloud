package loadbalancer_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/e2etests"
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
		Type:        e2etests.TestLoadBalancerType,
		NetworkZone: "eu-central",
	}

	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
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
		Type:        e2etests.TestLoadBalancerType,
		NetworkZone: "eu-central",
	}

	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
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

func TestAccHcloudLoadBalancerNetwork_CannotAttachToTwoNetworks(t *testing.T) {
	nwRess := make([]*network.RData, 2)
	snRess := make([]*network.RDataSubnet, len(nwRess))
	for i := 0; i < len(nwRess); i++ {
		nwName := fmt.Sprintf("test-network-%d", i)
		nwRes := &network.RData{Name: nwName, IPRange: "10.0.0.0/16"}
		nwRes.SetRName(nwName)
		nwRess[i] = nwRes

		snRes := &network.RDataSubnet{
			Type:        "cloud",
			NetworkID:   nwRes.TFID() + ".id",
			NetworkZone: "eu-central",
			IPRange:     "10.0.1.0/24",
		}
		snRes.SetRName(fmt.Sprintf("test-network-subnet-%d", i))
		snRess[i] = snRes
	}

	lbRes := &loadbalancer.RData{
		Name:        "lb-double-attach-test",
		Type:        e2etests.TestLoadBalancerType,
		NetworkZone: "eu-central",
	}

	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_network", nwRess[0],
					"testdata/r/hcloud_network", nwRess[1],
					"testdata/r/hcloud_network_subnet", snRess[0],
					"testdata/r/hcloud_network_subnet", snRess[1],
					"testdata/r/hcloud_load_balancer", lbRes,
					"testdata/r/hcloud_load_balancer_network", &loadbalancer.RDataNetwork{
						Name:           "test-network-0",
						LoadBalancerID: lbRes.TFID() + ".id",
						SubNetID:       snRess[0].TFID() + ".id",
						IP:             "10.0.1.5",
					},
					"testdata/r/hcloud_load_balancer_network", &loadbalancer.RDataNetwork{
						Name:           "test-network-1",
						LoadBalancerID: lbRes.TFID() + ".id",
						SubNetID:       snRess[1].TFID() + ".id",
						IP:             "10.0.1.5",
					},
				),
				ExpectError: regexp.MustCompile(`.*load_balancer_already_attached.*`),
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
