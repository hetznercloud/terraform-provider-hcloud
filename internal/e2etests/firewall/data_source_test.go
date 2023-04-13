package firewall

import (
	"fmt"
	"testing"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/e2etests"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/firewall"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccHcloudDataSourceFirewallTest(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	res := firewall.NewRData(t, "basic-firewall", []firewall.RDataRule{}, nil)
	res.SetRName("firewall-ds-test")
	firewallByName := &firewall.DData{
		FirewallName: res.TFID() + ".name",
	}
	firewallByName.SetRName("firewall_by_name")
	firewallByID := &firewall.DData{
		FirewallID: res.TFID() + ".id",
	}
	firewallByID.SetRName("firewall_by_id")
	firewallBySel := &firewall.DData{
		LabelSelector: fmt.Sprintf("key=${%s.labels[\"key\"]}", res.TFID()),
	}
	firewallBySel.SetRName("firewall_by_sel")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(firewall.ResourceType, firewall.ByID(t, nil)),
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
}

func TestAccHcloudDataSourceFirewallListTest(t *testing.T) {
	res := firewall.NewRData(t, "firewall-ds-test", []firewall.RDataRule{}, nil)

	firewallBySel := &firewall.DDataList{
		LabelSelector: fmt.Sprintf("key=${%s.labels[\"key\"]}", res.TFID()),
	}
	firewallBySel.SetRName("firewall_by_sel")

	allFirewallsSel := &firewall.DDataList{}
	allFirewallsSel.SetRName("all_firewalls_sel")

	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(firewall.ResourceType, firewall.ByID(t, nil)),
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
}
