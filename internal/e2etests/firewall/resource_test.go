package firewall

import (
	"fmt"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/e2etests"
	"testing"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/firewall"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestFirewallResource_Basic(t *testing.T) {
	var sk hcloud.Firewall

	res := firewall.NewRData(t, "basic-firewall", []firewall.RDataRule{
		{
			Direction: "in",
			Protocol:  "tcp",
			SourceIPs: []string{"0.0.0.0/0", "::/0"},
			Port:      "80",
		},
	})

	updated := firewall.NewRData(t, "basic-firewall", []firewall.RDataRule{
		{
			Direction: "in",
			Protocol:  "tcp",
			SourceIPs: []string{"0.0.0.0/0", "::/0"},
			Port:      "443",
		},
	})
	updated.SetRName(res.RName())
	tmplMan := testtemplate.Manager{}
	resource.Test(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(firewall.ResourceType, firewall.ByID(t, &sk)),
		Steps: []resource.TestStep{
			{
				// Create a new SSH Key using the required values
				// only.
				Config: tmplMan.Render(t, "testdata/r/hcloud_firewall", res),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), firewall.ByID(t, &sk)),
					resource.TestCheckResourceAttr(res.TFID(), "name",
						fmt.Sprintf("basic-firewall--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res.TFID(), "rule.#", "1"),
					resource.TestCheckResourceAttr(res.TFID(), "rule.0.direction", "in"),
					resource.TestCheckResourceAttr(res.TFID(), "rule.0.protocol", "tcp"),
					resource.TestCheckResourceAttr(res.TFID(), "rule.0.port", "80"),
					resource.TestCheckResourceAttr(res.TFID(), "rule.0.source_ips.#", "2"),
				),
			},
			{
				// Try to import the newly created SSH Key
				ResourceName:      res.TFID(),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Create a new SSH Key using the required values
				// only.
				Config: tmplMan.Render(t, "testdata/r/hcloud_firewall", updated),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), firewall.ByID(t, &sk)),
					resource.TestCheckResourceAttr(res.TFID(), "name",
						fmt.Sprintf("basic-firewall--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res.TFID(), "rule.#", "1"),
					resource.TestCheckResourceAttr(res.TFID(), "rule.0.direction", "in"),
					resource.TestCheckResourceAttr(res.TFID(), "rule.0.protocol", "tcp"),
					resource.TestCheckResourceAttr(res.TFID(), "rule.0.port", "443"),
					resource.TestCheckResourceAttr(res.TFID(), "rule.0.source_ips.#", "2"),
				),
			},
		},
	})
}
