package network

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	loadbalancertest "github.com/hetznercloud/terraform-provider-hcloud/test/e2e/loadbalancer"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/loadbalancer"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestNetworkDataSource(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		tmplMan := testtemplate.Manager{}

		res := &RData{
			Name:    "network-ds-test",
			IPRange: "10.0.0.0/16",
			Labels: map[string]string{
				"key": strconv.Itoa(acctest.RandInt()),
			},
			ExposeRoutesToVSwitch: true,
		}
		res.SetRName("network-ds-test")
		networkByName := &DData{
			NetworkName: res.TFID() + ".name",
		}
		networkByName.SetRName("network_by_name")
		networkByID := &DData{
			NetworkID: res.TFID() + ".id",
		}
		networkByID.SetRName("network_by_id")
		networkBySel := &DData{
			LabelSelector: fmt.Sprintf("key=${%s.labels[\"key\"]}", res.TFID()),
		}
		networkBySel.SetRName("network_by_sel")

		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 teste2e.PreCheck(t),
			ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
			CheckDestroy:             testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancertest.ByID(t, nil)),
			Steps: []resource.TestStep{
				{
					Config: tmplMan.Render(t,
						"testdata/r/hcloud_network", res,
					),
				},
				{
					Config: tmplMan.Render(t,
						"testdata/r/hcloud_network", res,
						"testdata/d/hcloud_network", networkByName,
						"testdata/d/hcloud_network", networkByID,
						"testdata/d/hcloud_network", networkBySel,
					),

					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(networkByName.TFID(),
							"name", fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt)),
						resource.TestCheckResourceAttr(networkByName.TFID(), "ip_range", res.IPRange),
						resource.TestCheckResourceAttr(networkByName.TFID(), "expose_routes_to_vswitch", "true"),

						resource.TestCheckResourceAttr(networkByID.TFID(),
							"name", fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt)),
						resource.TestCheckResourceAttr(networkByID.TFID(), "ip_range", res.IPRange),
						resource.TestCheckResourceAttr(networkByID.TFID(), "expose_routes_to_vswitch", "true"),

						resource.TestCheckResourceAttr(networkBySel.TFID(),
							"name", fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt)),
						resource.TestCheckResourceAttr(networkBySel.TFID(), "ip_range", res.IPRange),
						resource.TestCheckResourceAttr(networkBySel.TFID(), "expose_routes_to_vswitch", "true"),
					),
				},
			},
		})
	})

	t.Run("list", func(t *testing.T) {
		res := &RData{
			Name:    "network-ds-test",
			IPRange: "10.0.0.0/16",
			Labels: map[string]string{
				"key": strconv.Itoa(acctest.RandInt()),
			},
		}
		res.SetRName("network-ds-test")

		networksBySel := &DDataList{
			LabelSelector: fmt.Sprintf("key=${%s.labels[\"key\"]}", res.TFID()),
		}
		networksBySel.SetRName("networks_by_sel")

		allNetworksSel := &DDataList{}
		allNetworksSel.SetRName("all_networks_sel")

		tmplMan := testtemplate.Manager{}
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 teste2e.PreCheck(t),
			ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
			CheckDestroy:             testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancertest.ByID(t, nil)),
			Steps: []resource.TestStep{
				{
					Config: tmplMan.Render(t,
						"testdata/r/hcloud_network", res,
					),
				},
				{
					Config: tmplMan.Render(t,
						"testdata/r/hcloud_network", res,
						"testdata/d/hcloud_networks", networksBySel,
						"testdata/d/hcloud_networks", allNetworksSel,
					),

					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckTypeSetElemNestedAttrs(networksBySel.TFID(), "networks.*",
							map[string]string{
								"name": fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt),
							},
						),

						resource.TestCheckTypeSetElemNestedAttrs(allNetworksSel.TFID(), "networks.*",
							map[string]string{
								"name": fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt),
							},
						),
					),
				},
			},
		})
	})
}
