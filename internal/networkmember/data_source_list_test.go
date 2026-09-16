package networkmember_test

import (
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/network"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/networkmember"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/primaryip"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/server"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testmux"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccNetworkMemberDataSourceList(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	ntws := network.NewBlueprint(t)
	srvs := server.NewBlueprint(t)

	res := &server.RDataNetwork{
		Name:     "attachment",
		ServerID: srvs.ServerA.TFID() + ".id",
		SubNetID: ntws.SubnetA1.TFID() + ".id",
		IP:       "10.0.1.5",
		AliasIPs: []string{"10.0.1.6", "10.0.1.7"},
	}
	res.SetRName("attachment")

	byType := &networkmember.DDataList{
		NetworkID: ntws.NetworkA.TFID() + ".id",
		WithType:  []string{"server"},
	}
	byType.SetRName("by_type")

	byTypeEmpty := &networkmember.DDataList{
		NetworkID: ntws.NetworkA.TFID() + ".id",
		WithType:  []string{"load_balancer"},
	}
	byTypeEmpty.SetRName("by_type_empty")

	byStatus := &networkmember.DDataList{
		NetworkID:  ntws.NetworkA.TFID() + ".id",
		WithStatus: []string{"ok"},
	}
	byStatus.SetRName("by_status")

	byStatusEmpty := &networkmember.DDataList{
		NetworkID:  ntws.NetworkA.TFID() + ".id",
		WithStatus: []string{"error"},
	}
	byStatusEmpty.SetRName("by_status_empty")

	bySubnet := &networkmember.DDataList{
		NetworkID:  ntws.NetworkA.TFID() + ".id",
		WithSubnet: []string{ntws.SubnetA1.IPRange},
	}
	bySubnet.SetRName("by_subnet")

	bySubnetEmpty := &networkmember.DDataList{
		NetworkID:  ntws.NetworkA.TFID() + ".id",
		WithSubnet: []string{ntws.SubnetA2.IPRange},
	}
	bySubnetEmpty.SetRName("by_subnet_empty")

	all := &networkmember.DDataList{
		NetworkID: ntws.NetworkA.TFID() + ".id",
	}
	all.SetRName("all")

	attributesStateChecks := func(tfid string) []statecheck.StateCheck {
		return []statecheck.StateCheck{
			statecheck.ExpectKnownValue(tfid, tfjsonpath.New("members"), knownvalue.ListSizeExact(1)),
			statecheck.ExpectKnownValue(tfid, tfjsonpath.New("members").AtSliceIndex(0).AtMapKey("id"), knownvalue.NotNull()),
			statecheck.ExpectKnownValue(tfid, tfjsonpath.New("members").AtSliceIndex(0).AtMapKey("type"), knownvalue.StringExact("server")),
			statecheck.ExpectKnownValue(tfid, tfjsonpath.New("members").AtSliceIndex(0).AtMapKey("ip"), knownvalue.StringExact(res.IP)),
			statecheck.ExpectKnownValue(tfid, tfjsonpath.New("members").AtSliceIndex(0).AtMapKey("alias_ips"), knownvalue.SetExact([]knownvalue.Check{knownvalue.StringExact(res.AliasIPs[0]), knownvalue.StringExact(res.AliasIPs[1])})),
			statecheck.ExpectKnownValue(tfid, tfjsonpath.New("members").AtSliceIndex(0).AtMapKey("subnet"), knownvalue.StringExact(ntws.SubnetA1.IPRange)),
			statecheck.ExpectKnownValue(tfid, tfjsonpath.New("members").AtSliceIndex(0).AtMapKey("status"), knownvalue.StringExact("ok")),
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(primaryip.ResourceType, primaryip.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_network", ntws.NetworkA,
					"testdata/r/hcloud_network_subnet", ntws.SubnetA1,
					"testdata/r/hcloud_network_subnet", ntws.SubnetA2,
					"testdata/r/hcloud_server", srvs.ServerA,
					"testdata/r/hcloud_server_network", res,
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_network", ntws.NetworkA,
					"testdata/r/hcloud_network_subnet", ntws.SubnetA1,
					"testdata/r/hcloud_network_subnet", ntws.SubnetA2,
					"testdata/r/hcloud_server", srvs.ServerA,
					"testdata/r/hcloud_server_network", res,

					"testdata/d/hcloud_network_members", byType,
					"testdata/d/hcloud_network_members", byTypeEmpty,
					"testdata/d/hcloud_network_members", byStatus,
					"testdata/d/hcloud_network_members", byStatusEmpty,
					"testdata/d/hcloud_network_members", bySubnet,
					"testdata/d/hcloud_network_members", bySubnetEmpty,
					"testdata/d/hcloud_network_members", all,
				),
				ConfigStateChecks: slices.Concat([]statecheck.StateCheck{
					statecheck.ExpectKnownValue(byTypeEmpty.TFID(), tfjsonpath.New("members"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(byStatusEmpty.TFID(), tfjsonpath.New("members"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(bySubnetEmpty.TFID(), tfjsonpath.New("members"), knownvalue.ListSizeExact(0)),
				},
					attributesStateChecks(byType.TFID()),
					attributesStateChecks(byStatus.TFID()),
					attributesStateChecks(bySubnet.TFID()),
					attributesStateChecks(all.TFID()),
				),
			},
		},
	})
}
