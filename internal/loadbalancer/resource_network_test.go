package loadbalancer_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/loadbalancer"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/network"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testmux"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccLoadBalancerNetworkResource_NetworkID(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	var (
		hcNetwork      hcloud.Network
		hcLoadBalancer hcloud.LoadBalancer
	)

	ntws := network.NewBlueprint(t)
	lbls := loadbalancer.NewBlueprint(t)

	res1 := &loadbalancer.RDataNetwork{
		Name:                  "attachment",
		LoadBalancerID:        lbls.LoadBalancerA.TFID() + ".id",
		NetworkID:             ntws.NetworkA.TFID() + ".id",
		IP:                    "10.0.1.5",
		EnablePublicInterface: new(false),
		DependsOn:             []string{ntws.SubnetA1.TFID()},
	}
	res1.SetRName("attachment")

	res2 := &loadbalancer.RDataNetwork{
		Name:                  res1.Name,
		LoadBalancerID:        res1.LoadBalancerID,
		NetworkID:             res1.NetworkID,
		IP:                    res1.IP,
		EnablePublicInterface: new(true),
		DependsOn:             res1.DependsOn,
	}
	res2.SetRName("attachment")

	res3 := &loadbalancer.RDataNetwork{
		Name:           res1.Name,
		LoadBalancerID: res1.LoadBalancerID,
		NetworkID:      res1.NetworkID,
		DependsOn:      res1.DependsOn,
	}
	res3.SetRName("attachment")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_network", ntws.NetworkA,
					"testdata/r/hcloud_network_subnet", ntws.SubnetA1,
					"testdata/r/hcloud_network_subnet", ntws.SubnetA2,
					"testdata/r/hcloud_load_balancer", lbls.LoadBalancerA,
					"testdata/r/hcloud_load_balancer_network", res1,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(ntws.NetworkA.TFID(), network.ByID(t, &hcNetwork)),
					testsupport.CheckResourceExists(lbls.LoadBalancerA.TFID(), loadbalancer.ByID(t, &hcLoadBalancer)),
					testsupport.LiftTCF(hasLoadBalancerNetwork(t, &hcLoadBalancer, &hcNetwork, "10.0.1.5")),
					resource.TestCheckResourceAttr(res1.TFID(), "ip", "10.0.1.5"),
					resource.TestCheckResourceAttr(res1.TFID(), "enable_public_interface", "false"),
				),
			},
			{
				ResourceName:      res1.TFID(),
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(_ *terraform.State) (string, error) {
					return fmt.Sprintf("%d-%d", hcLoadBalancer.ID, hcNetwork.ID), nil
				},
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_network", ntws.NetworkA,
					"testdata/r/hcloud_network_subnet", ntws.SubnetA1,
					"testdata/r/hcloud_network_subnet", ntws.SubnetA2,
					"testdata/r/hcloud_load_balancer", lbls.LoadBalancerA,
					"testdata/r/hcloud_load_balancer_network", res2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(res1.TFID(), "ip", "10.0.1.5"),
					resource.TestCheckResourceAttr(res1.TFID(), "enable_public_interface", "true"),
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_network", ntws.NetworkA,
					"testdata/r/hcloud_network_subnet", ntws.SubnetA1,
					"testdata/r/hcloud_network_subnet", ntws.SubnetA2,
					"testdata/r/hcloud_load_balancer", lbls.LoadBalancerA,
					"testdata/r/hcloud_load_balancer_network", res3,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(res1.TFID(), "ip", "10.0.1.5"),
					resource.TestCheckResourceAttr(res1.TFID(), "enable_public_interface", "true"),
				),
			},
		},
	})
}

func TestAccLoadBalancerNetworkResource_SubnetID(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	var (
		hcNetwork      hcloud.Network
		hcLoadBalancer hcloud.LoadBalancer
	)

	ntws := network.NewBlueprint(t)
	lbls := loadbalancer.NewBlueprint(t)

	res1 := &loadbalancer.RDataNetwork{
		Name:           "attachment",
		LoadBalancerID: lbls.LoadBalancerA.TFID() + ".id",
		SubNetID:       ntws.SubnetA1.TFID() + ".id",
	}
	res1.SetRName("attachment")

	res2 := &loadbalancer.RDataNetwork{
		Name:           res1.Name,
		LoadBalancerID: res1.LoadBalancerID,
		SubNetID:       res1.SubNetID,
		IP:             "10.0.1.5",
	}
	res2.SetRName("attachment")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_network", ntws.NetworkA,
					"testdata/r/hcloud_network_subnet", ntws.SubnetA1,
					"testdata/r/hcloud_network_subnet", ntws.SubnetA2,
					"testdata/r/hcloud_load_balancer", lbls.LoadBalancerA,
					"testdata/r/hcloud_load_balancer_network", res1,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(ntws.NetworkA.TFID(), network.ByID(t, &hcNetwork)),
					testsupport.CheckResourceExists(lbls.LoadBalancerA.TFID(), loadbalancer.ByID(t, &hcLoadBalancer)),
					testsupport.LiftTCF(hasLoadBalancerNetwork(t, &hcLoadBalancer, &hcNetwork, "10.0.1.1")),
					resource.TestCheckResourceAttr(res1.TFID(), "ip", "10.0.1.1"),
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_network", ntws.NetworkA,
					"testdata/r/hcloud_network_subnet", ntws.SubnetA1,
					"testdata/r/hcloud_network_subnet", ntws.SubnetA2,
					"testdata/r/hcloud_load_balancer", lbls.LoadBalancerA,
					"testdata/r/hcloud_load_balancer_network", res2,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(ntws.NetworkA.TFID(), network.ByID(t, &hcNetwork)),
					testsupport.CheckResourceExists(lbls.LoadBalancerA.TFID(), loadbalancer.ByID(t, &hcLoadBalancer)),
					testsupport.LiftTCF(hasLoadBalancerNetwork(t, &hcLoadBalancer, &hcNetwork, "10.0.1.5")),
					resource.TestCheckResourceAttr(res2.TFID(), "ip", "10.0.1.5"),
				),
			},
		},
	})
}

func TestAccLoadBalancerNetworkResource_CannotAttachToTwoNetworks(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	ntws := network.NewBlueprint(t)
	lbls := loadbalancer.NewBlueprint(t)

	res1 := &loadbalancer.RDataNetwork{
		Name:           "attachment1",
		LoadBalancerID: lbls.LoadBalancerA.TFID() + ".id",
		SubNetID:       ntws.SubnetA1.TFID() + ".id",
	}
	res1.SetRName(res1.Name)

	res2 := &loadbalancer.RDataNetwork{
		Name:           "attachment2",
		LoadBalancerID: lbls.LoadBalancerA.TFID() + ".id",
		SubNetID:       ntws.SubnetB1.TFID() + ".id",
	}
	res2.SetRName(res2.Name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_network", ntws.NetworkA,
					"testdata/r/hcloud_network", ntws.NetworkB,
					"testdata/r/hcloud_network_subnet", ntws.SubnetA1,
					"testdata/r/hcloud_network_subnet", ntws.SubnetB1,
					"testdata/r/hcloud_load_balancer", lbls.LoadBalancerA,
					"testdata/r/hcloud_load_balancer_network", res1,
					"testdata/r/hcloud_load_balancer_network", res2,
				),
				ExpectError: regexp.MustCompile(`.*load_balancer_already_attached.*`),
			},
		},
	})
}

func hasLoadBalancerNetwork(t *testing.T, lb *hcloud.LoadBalancer, nw *hcloud.Network, ips ...string) func() error {
	return func() error {
		attachment := lb.PrivateNetFor(nw)
		if !assert.NotNil(t, attachment, "load balancer has no private network") {
			return nil
		}
		assert.Contains(t, ips, attachment.IP.String())
		return nil
	}
}
