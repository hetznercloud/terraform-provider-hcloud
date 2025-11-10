package loadbalancer_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/assert"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/loadbalancer"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/network"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

type LoadBalancerNetworkBlueprint struct {
	prefixes []string

	network *network.RData
	subnet1 *network.RDataSubnet
	subnet2 *network.RDataSubnet

	loadBalancer1 *loadbalancer.RData
	loadBalancer2 *loadbalancer.RData
}

func (b *LoadBalancerNetworkBlueprint) HCName(name string) string {
	return strings.Join(append(b.prefixes, name), "-")
}

func (b *LoadBalancerNetworkBlueprint) TFName(name string) string {
	return strings.Join(append(b.prefixes, name), "_")
}

func makeLoadBalancerNetworkBlueprint(t *testing.T, prefixes ...string) *LoadBalancerNetworkBlueprint {
	t.Helper()

	b := &LoadBalancerNetworkBlueprint{prefixes: prefixes}

	b.network = &network.RData{
		Name:    b.HCName("network"),
		IPRange: "10.0.0.0/16",
	}
	b.network.SetRName(b.TFName("network"))

	b.subnet1 = &network.RDataSubnet{
		NetworkID:   b.network.TFID() + ".id",
		NetworkZone: "eu-central",
		IPRange:     "10.0.1.0/24",
		Type:        "cloud",
	}
	b.subnet1.SetRName(b.TFName("subnet1"))

	b.subnet2 = &network.RDataSubnet{
		NetworkID:   b.network.TFID() + ".id",
		NetworkZone: "eu-central",
		IPRange:     "10.0.2.0/24",
		Type:        "cloud",
	}
	b.subnet2.SetRName(b.TFName("subnet2"))

	b.loadBalancer1 = &loadbalancer.RData{
		Name:        b.HCName("loadbalancer1"),
		Type:        teste2e.TestLoadBalancerType,
		NetworkZone: "eu-central",
	}
	b.loadBalancer1.SetRName(b.TFName("loadbalancer1"))

	b.loadBalancer2 = &loadbalancer.RData{
		Name:        b.HCName("loadbalancer2"),
		Type:        teste2e.TestLoadBalancerType,
		NetworkZone: "eu-central",
	}
	b.loadBalancer2.SetRName(b.TFName("loadbalancer2"))

	return b
}

func TestAccLoadBalancerNetworkResource_NetworkID(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	var (
		hcNetwork      hcloud.Network
		hcLoadBalancer hcloud.LoadBalancer
	)

	b := makeLoadBalancerNetworkBlueprint(t)

	res1 := &loadbalancer.RDataNetwork{
		Name:                  "attachment",
		LoadBalancerID:        b.loadBalancer1.TFID() + ".id",
		NetworkID:             b.network.TFID() + ".id",
		IP:                    "10.0.1.5",
		EnablePublicInterface: hcloud.Ptr(false),
		DependsOn:             []string{b.subnet1.TFID()},
	}
	res1.SetRName("attachment")

	res2 := &loadbalancer.RDataNetwork{
		Name:                  res1.Name,
		LoadBalancerID:        res1.LoadBalancerID,
		NetworkID:             res1.NetworkID,
		IP:                    res1.IP,
		EnablePublicInterface: hcloud.Ptr(true),
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
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_network", b.network,
					"testdata/r/hcloud_network_subnet", b.subnet1,
					"testdata/r/hcloud_load_balancer", b.loadBalancer1,
					"testdata/r/hcloud_load_balancer_network", res1,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(b.network.TFID(), network.ByID(t, &hcNetwork)),
					testsupport.CheckResourceExists(b.loadBalancer1.TFID(), loadbalancer.ByID(t, &hcLoadBalancer)),
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
					"testdata/r/hcloud_network", b.network,
					"testdata/r/hcloud_network_subnet", b.subnet1,
					"testdata/r/hcloud_load_balancer", b.loadBalancer1,
					"testdata/r/hcloud_load_balancer_network", res2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(res1.TFID(), "ip", "10.0.1.5"),
					resource.TestCheckResourceAttr(res1.TFID(), "enable_public_interface", "true"),
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_network", b.network,
					"testdata/r/hcloud_network_subnet", b.subnet1,
					"testdata/r/hcloud_load_balancer", b.loadBalancer1,
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

	b := makeLoadBalancerNetworkBlueprint(t)

	res1 := &loadbalancer.RDataNetwork{
		Name:           "attachment",
		LoadBalancerID: b.loadBalancer1.TFID() + ".id",
		SubNetID:       b.subnet1.TFID() + ".id",
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
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_network", b.network,
					"testdata/r/hcloud_network_subnet", b.subnet1,
					"testdata/r/hcloud_load_balancer", b.loadBalancer1,
					"testdata/r/hcloud_load_balancer_network", res1,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(b.network.TFID(), network.ByID(t, &hcNetwork)),
					testsupport.CheckResourceExists(b.loadBalancer1.TFID(), loadbalancer.ByID(t, &hcLoadBalancer)),
					testsupport.LiftTCF(hasLoadBalancerNetwork(t, &hcLoadBalancer, &hcNetwork, "10.0.1.1")),
					resource.TestCheckResourceAttr(res1.TFID(), "ip", "10.0.1.1"),
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_network", b.network,
					"testdata/r/hcloud_network_subnet", b.subnet1,
					"testdata/r/hcloud_load_balancer", b.loadBalancer1,
					"testdata/r/hcloud_load_balancer_network", res2,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(b.network.TFID(), network.ByID(t, &hcNetwork)),
					testsupport.CheckResourceExists(b.loadBalancer1.TFID(), loadbalancer.ByID(t, &hcLoadBalancer)),
					testsupport.LiftTCF(hasLoadBalancerNetwork(t, &hcLoadBalancer, &hcNetwork, "10.0.1.5")),
					resource.TestCheckResourceAttr(res2.TFID(), "ip", "10.0.1.5"),
				),
			},
		},
	})
}

func TestAccLoadBalancerNetworkResource_CannotAttachToTwoNetworks(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	b1 := makeLoadBalancerNetworkBlueprint(t, "b1")
	b2 := makeLoadBalancerNetworkBlueprint(t, "b2")

	res1 := &loadbalancer.RDataNetwork{
		Name:           "attachment1",
		LoadBalancerID: b1.loadBalancer1.TFID() + ".id",
		SubNetID:       b1.subnet1.TFID() + ".id",
	}
	res1.SetRName(res1.Name)

	res2 := &loadbalancer.RDataNetwork{
		Name:           "attachment2",
		LoadBalancerID: b1.loadBalancer1.TFID() + ".id",
		SubNetID:       b2.subnet1.TFID() + ".id",
	}
	res2.SetRName(res2.Name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_network", b1.network,
					"testdata/r/hcloud_network", b2.network,
					"testdata/r/hcloud_network_subnet", b1.subnet1,
					"testdata/r/hcloud_network_subnet", b2.subnet1,
					"testdata/r/hcloud_load_balancer", b1.loadBalancer1,
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

func TestAccLoadBalancerNetworkResource_UpgradePluginFramework(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	b := makeLoadBalancerNetworkBlueprint(t)

	res := &loadbalancer.RDataNetwork{
		Name:           "attachment",
		LoadBalancerID: b.loadBalancer1.TFID() + ".id",
		SubNetID:       b.subnet2.TFID() + ".id",
	}
	res.SetRName("attachment")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: teste2e.PreCheck(t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"hcloud": {
						VersionConstraint: "1.54.0",
						Source:            "hetznercloud/hcloud",
					},
				},
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_network", b.network,
					"testdata/r/hcloud_network_subnet", b.subnet1,
					"testdata/r/hcloud_network_subnet", b.subnet2,
					"testdata/r/hcloud_load_balancer", b.loadBalancer1,
					"testdata/r/hcloud_load_balancer_network", res,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(res.TFID(),
						tfjsonpath.New("ip"),
						knownvalue.StringExact("10.0.2.1")),
					statecheck.ExpectKnownValue(res.TFID(),
						tfjsonpath.New("enable_public_interface"),
						knownvalue.Bool(true)),
				},
			},
			{
				ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),

				Config: tmplMan.Render(t,
					"testdata/r/hcloud_network", b.network,
					"testdata/r/hcloud_network_subnet", b.subnet1,
					"testdata/r/hcloud_network_subnet", b.subnet2,
					"testdata/r/hcloud_load_balancer", b.loadBalancer1,
					"testdata/r/hcloud_load_balancer_network", res,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(res.TFID(),
						tfjsonpath.New("ip"),
						knownvalue.StringExact("10.0.2.1")),
					statecheck.ExpectKnownValue(res.TFID(),
						tfjsonpath.New("enable_public_interface"),
						knownvalue.Bool(true)),
				},
			},
		},
	})
}
