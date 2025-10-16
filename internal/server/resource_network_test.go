package server_test

import (
	"fmt"
	"net"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/assert"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/network"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/server"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

type ServerNetworkBlueprint struct {
	network *network.RData
	subnet1 *network.RDataSubnet
	subnet2 *network.RDataSubnet

	server1 *server.RData
	server2 *server.RData
}

func makeServerNetworkBlueprint(t *testing.T) *ServerNetworkBlueprint {
	t.Helper()

	b := &ServerNetworkBlueprint{}

	b.network = &network.RData{
		Name:    "network",
		IPRange: "10.0.0.0/16",
	}
	b.network.SetRName("network")

	b.subnet1 = &network.RDataSubnet{
		NetworkID:   b.network.TFID() + ".id",
		NetworkZone: "eu-central",
		IPRange:     "10.0.1.0/24",
		Type:        "cloud",
	}
	b.subnet1.SetRName("subnet1")

	b.subnet2 = &network.RDataSubnet{
		NetworkID:   b.network.TFID() + ".id",
		NetworkZone: "eu-central",
		IPRange:     "10.0.2.0/24",
		Type:        "cloud",

		// Ensures subnet 2 is created after subnet 1
		DependsOn: []string{b.subnet1.TFID()},
	}
	b.subnet2.SetRName("subnet2")

	b.server1 = &server.RData{
		Name:       "server1",
		Type:       teste2e.TestServerType,
		Datacenter: teste2e.TestDataCenter,
		Image:      teste2e.TestImage,
	}
	b.server1.SetRName("server1")

	b.server2 = &server.RData{
		Name:       "server2",
		Type:       teste2e.TestServerType,
		Datacenter: teste2e.TestDataCenter,
		Image:      teste2e.TestImage,
	}
	b.server2.SetRName("server2")

	return b
}

func TestAccServerNetworkResource_NetworkID(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	var (
		hcNetwork hcloud.Network
		hcServer  hcloud.Server
	)

	b := makeServerNetworkBlueprint(t)

	res1 := &server.RDataNetwork{
		Name:      "attachment",
		ServerID:  b.server1.TFID() + ".id",
		NetworkID: b.network.TFID() + ".id",
		IP:        "10.0.1.5",
		AliasIPs:  []string{"10.0.1.6", "10.0.1.7"},
		DependsOn: []string{b.subnet1.TFID()},
	}
	res1.SetRName("attachment")

	// Remove alias ips
	res2 := &server.RDataNetwork{
		Name:      res1.Name,
		ServerID:  res1.ServerID,
		NetworkID: res1.NetworkID,
		IP:        res1.IP,
		DependsOn: res1.DependsOn,
	}
	res2.SetRName("attachment")

	// Add other alias ips
	res3 := &server.RDataNetwork{
		Name:      res1.Name,
		ServerID:  res1.ServerID,
		NetworkID: res1.NetworkID,
		IP:        res1.IP,
		AliasIPs:  []string{"10.0.1.8"},
		DependsOn: res1.DependsOn,
	}
	res3.SetRName("attachment")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_network", b.network,
					"testdata/r/hcloud_network_subnet", b.subnet1,
					"testdata/r/hcloud_network_subnet", b.subnet2,
					"testdata/r/hcloud_server", b.server1,
					"testdata/r/hcloud_server_network", res1,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(b.server1.TFID(), server.ByID(t, &hcServer)),
					testsupport.CheckResourceExists(b.network.TFID(), network.ByID(t, &hcNetwork)),
					testsupport.LiftTCF(hasServerNetwork(t, &hcServer, &hcNetwork, "10.0.1.5")),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(res1.TFID(),
						tfjsonpath.New("ip"),
						knownvalue.StringExact("10.0.1.5")),
					statecheck.ExpectKnownValue(res1.TFID(),
						tfjsonpath.New("alias_ips"),
						knownvalue.SetExact([]knownvalue.Check{
							knownvalue.StringExact("10.0.1.6"),
							knownvalue.StringExact("10.0.1.7"),
						})),
					statecheck.ExpectKnownValue(res1.TFID(),
						tfjsonpath.New("mac_address"),
						knownvalue.StringFunc(func(v string) error {
							_, err := net.ParseMAC(v)
							return err
						})),
				},
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_network", b.network,
					"testdata/r/hcloud_network_subnet", b.subnet1,
					"testdata/r/hcloud_network_subnet", b.subnet2,
					"testdata/r/hcloud_server", b.server1,
					"testdata/r/hcloud_server_network", res2,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(b.server1.TFID(), server.ByID(t, &hcServer)),
					testsupport.CheckResourceExists(b.network.TFID(), network.ByID(t, &hcNetwork)),
					testsupport.LiftTCF(hasServerNetwork(t, &hcServer, &hcNetwork, "10.0.1.5")),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(res1.TFID(),
						tfjsonpath.New("ip"),
						knownvalue.StringExact("10.0.1.5")),
					statecheck.ExpectKnownValue(res1.TFID(),
						tfjsonpath.New("alias_ips"),
						knownvalue.SetSizeExact(0)),
					statecheck.ExpectKnownValue(res1.TFID(),
						tfjsonpath.New("mac_address"),
						knownvalue.StringFunc(func(v string) error {
							_, err := net.ParseMAC(v)
							return err
						})),
				},
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_network", b.network,
					"testdata/r/hcloud_network_subnet", b.subnet1,
					"testdata/r/hcloud_network_subnet", b.subnet2,
					"testdata/r/hcloud_server", b.server1,
					"testdata/r/hcloud_server_network", res3,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(b.server1.TFID(), server.ByID(t, &hcServer)),
					testsupport.CheckResourceExists(b.network.TFID(), network.ByID(t, &hcNetwork)),
					testsupport.LiftTCF(hasServerNetwork(t, &hcServer, &hcNetwork, "10.0.1.5")),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(res1.TFID(),
						tfjsonpath.New("ip"),
						knownvalue.StringExact("10.0.1.5")),
					statecheck.ExpectKnownValue(res1.TFID(),
						tfjsonpath.New("alias_ips"),
						knownvalue.SetExact([]knownvalue.Check{
							knownvalue.StringExact("10.0.1.8"),
						})),
					statecheck.ExpectKnownValue(res1.TFID(),
						tfjsonpath.New("mac_address"),
						knownvalue.StringFunc(func(v string) error {
							_, err := net.ParseMAC(v)
							return err
						})),
				},
			},
			{
				ResourceName:      res1.TFID(),
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(_ *terraform.State) (string, error) {
					return fmt.Sprintf("%d-%d", hcServer.ID, hcNetwork.ID), nil
				},
			},
		},
	})
}

func TestAccServerNetworkResource_SubnetID(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	var (
		hcNetwork hcloud.Network
		hcServer  hcloud.Server
	)

	b := makeServerNetworkBlueprint(t)

	res1 := &server.RDataNetwork{
		Name:     "attachment",
		ServerID: b.server1.TFID() + ".id",
		SubNetID: b.subnet2.TFID() + ".id",
	}
	res1.SetRName("attachment")

	// Remove alias ips
	res2 := &server.RDataNetwork{
		Name:     res1.Name,
		ServerID: res1.ServerID,
		SubNetID: res1.SubNetID,
	}
	res2.SetRName("attachment")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_network", b.network,
					"testdata/r/hcloud_network_subnet", b.subnet1,
					"testdata/r/hcloud_network_subnet", b.subnet2,
					"testdata/r/hcloud_server", b.server1,
					"testdata/r/hcloud_server_network", res1,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(b.server1.TFID(), server.ByID(t, &hcServer)),
					testsupport.CheckResourceExists(b.network.TFID(), network.ByID(t, &hcNetwork)),
					testsupport.LiftTCF(hasServerNetwork(t, &hcServer, &hcNetwork, "10.0.2.1")),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(res1.TFID(),
						tfjsonpath.New("ip"),
						knownvalue.StringExact("10.0.2.1")),
				},
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_network", b.network,
					"testdata/r/hcloud_network_subnet", b.subnet1,
					"testdata/r/hcloud_network_subnet", b.subnet2,
					"testdata/r/hcloud_server", b.server1,
					"testdata/r/hcloud_server_network", res2,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(b.server1.TFID(), server.ByID(t, &hcServer)),
					testsupport.CheckResourceExists(b.network.TFID(), network.ByID(t, &hcNetwork)),
					testsupport.LiftTCF(hasServerNetwork(t, &hcServer, &hcNetwork, "10.0.2.1")),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(res1.TFID(),
						tfjsonpath.New("ip"),
						knownvalue.StringExact("10.0.2.1")),
				},
			},
		},
	})
}

func hasServerNetwork(t *testing.T, s *hcloud.Server, nw *hcloud.Network, ips ...string) func() error {
	return func() error {
		var privNet *hcloud.ServerPrivateNet

		for _, n := range s.PrivateNet {
			if n.Network.ID == nw.ID {
				privNet = &n
				break
			}
		}
		if !assert.NotNil(t, privNet, "server has no private network") {
			return nil
		}
		assert.Contains(t, ips, privNet.IP.String())
		if len(ips) > 1 {
			for _, aip := range privNet.Aliases {
				assert.Contains(t, ips, aip.String())
			}
		}

		return nil
	}
}

func TestAccServerNetworkResource_UpgradePluginFramework(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	b := makeServerNetworkBlueprint(t)

	res := &server.RDataNetwork{
		Name:     "attachment",
		ServerID: b.server1.TFID() + ".id",
		SubNetID: b.subnet2.TFID() + ".id",
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
					"testdata/r/hcloud_server", b.server1,
					"testdata/r/hcloud_server_network", res,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(res.TFID(),
						tfjsonpath.New("ip"),
						// Attach to subnet using ip_range is not yet suported, and we
						// got assigned to subnet1
						knownvalue.StringExact("10.0.1.1")),
				},
			},
			{
				ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),

				Config: tmplMan.Render(t,
					"testdata/r/hcloud_network", b.network,
					"testdata/r/hcloud_network_subnet", b.subnet1,
					"testdata/r/hcloud_network_subnet", b.subnet2,
					"testdata/r/hcloud_server", b.server1,
					"testdata/r/hcloud_server_network", res,
				),

				PlanOnly: true,
			},
		},
	})
}

func TestAccServerNetworkResource_UpgradePluginFramework_AliasIPs(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	b := makeServerNetworkBlueprint(t)

	res := &server.RDataNetwork{
		Name:     "attachment",
		ServerID: b.server1.TFID() + ".id",
		SubNetID: b.subnet2.TFID() + ".id",
		AliasIPs: []string{"10.0.2.6", "10.0.2.7"},
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
					"testdata/r/hcloud_server", b.server1,
					"testdata/r/hcloud_server_network", res,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(res.TFID(),
						tfjsonpath.New("ip"),
						// Attach to subnet using ip_ranfe is not yet suported, and we
						// got assigned to subnet1
						knownvalue.StringExact("10.0.1.1")),
				},
			},
			{
				ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),

				Config: tmplMan.Render(t,
					"testdata/r/hcloud_network", b.network,
					"testdata/r/hcloud_network_subnet", b.subnet1,
					"testdata/r/hcloud_network_subnet", b.subnet2,
					"testdata/r/hcloud_server", b.server1,
					"testdata/r/hcloud_server_network", res,
				),

				PlanOnly: true,
			},
		},
	})
}
