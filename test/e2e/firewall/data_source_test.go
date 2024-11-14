package firewall

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/firewall"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestFirewallDataSource(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		tmplMan := testtemplate.Manager{}

		res := NewRData(t, "basic-firewall", []RDataRule{}, nil)
		res.SetRName("firewall-ds-test")
		firewallByName := &DData{
			FirewallName: res.TFID() + ".name",
		}
		firewallByName.SetRName("firewall_by_name")
		firewallByID := &DData{
			FirewallID: res.TFID() + ".id",
		}
		firewallByID.SetRName("firewall_by_id")
		firewallBySel := &DData{
			LabelSelector: fmt.Sprintf("key=${%s.labels[\"key\"]}", res.TFID()),
		}
		firewallBySel.SetRName("firewall_by_sel")

		// TODO: Move to parallel test once API endpoint supports higher parallelism
		resource.Test(t, resource.TestCase{
			PreCheck:                 teste2e.PreCheck(t),
			ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
			CheckDestroy:             testsupport.CheckResourcesDestroyed(firewall.ResourceType, ByID(t, nil)),
			Steps: []resource.TestStep{
				{
					Config: tmplMan.Render(t,
						"testdata/r/hcloud_firewall", res,
					),
				},
				{
					Config: tmplMan.Render(t,
						"testdata/r/hcloud_firewall", res,
						"testdata/d/hcloud_firewall", firewallByName,
						"testdata/d/hcloud_firewall", firewallByID,
						"testdata/d/hcloud_firewall", firewallBySel,
					),

					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(firewallByName.TFID(),
							"name", fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt)),

						resource.TestCheckResourceAttr(firewallByID.TFID(),
							"name", fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt)),

						resource.TestCheckResourceAttr(firewallBySel.TFID(),
							"name", fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt)),
					),
				},
			},
		})
	})

	t.Run("list", func(t *testing.T) {
		res := NewRData(t, "firewall-ds-test", []RDataRule{}, nil)

		firewallBySel := &DDataList{
			LabelSelector: fmt.Sprintf("key=${%s.labels[\"key\"]}", res.TFID()),
		}
		firewallBySel.SetRName("firewall_by_sel")

		allFirewallsSel := &DDataList{}
		allFirewallsSel.SetRName("all_firewalls_sel")

		tmplMan := testtemplate.Manager{}
		// TODO: Move to parallel test once API endpoint supports higher parallelism
		resource.Test(t, resource.TestCase{
			PreCheck:                 teste2e.PreCheck(t),
			ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
			CheckDestroy:             testsupport.CheckResourcesDestroyed(firewall.ResourceType, ByID(t, nil)),
			Steps: []resource.TestStep{
				{
					Config: tmplMan.Render(t,
						"testdata/r/hcloud_firewall", res,
					),
				},
				{
					Config: tmplMan.Render(t,
						"testdata/r/hcloud_firewall", res,
						"testdata/d/hcloud_firewalls", firewallBySel,
						"testdata/d/hcloud_firewalls", allFirewallsSel,
					),

					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckTypeSetElemNestedAttrs(firewallBySel.TFID(), "firewalls.*",
							map[string]string{
								"name": fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt),
							},
						),

						resource.TestCheckTypeSetElemNestedAttrs(allFirewallsSel.TFID(), "firewalls.*",
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
