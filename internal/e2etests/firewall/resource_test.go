package firewall

import (
	"fmt"
	"testing"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/e2etests"
	"github.com/stretchr/testify/assert"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/firewall"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestFirewallResource_Basic(t *testing.T) {
	var f hcloud.Firewall

	res := firewall.NewRData(t, "basic-firewall", []firewall.RDataRule{
		{
			Direction: "in",
			Protocol:  "tcp",
			SourceIPs: []string{"0.0.0.0/0", "::/0"},
			Port:      "80",
		},
		{
			Direction:      "out",
			Protocol:       "tcp",
			DestinationIPs: []string{"0.0.0.0/0", "::/0"},
			Port:           "80",
		},
		{
			Direction: "in",
			Protocol:  "udp",
			SourceIPs: []string{"0.0.0.0/0", "::/0"},
			Port:      "any",
		},
	})

	updated := firewall.NewRData(t, "basic-firewall", []firewall.RDataRule{
		{
			Direction: "in",
			Protocol:  "tcp",
			SourceIPs: []string{"0.0.0.0/0", "::/0"},
			Port:      "443",
		},
		{
			Direction:      "out",
			Protocol:       "tcp",
			DestinationIPs: []string{"0.0.0.0/0", "::/0"},
			Port:           "443",
		},
		{
			Direction: "in",
			Protocol:  "udp",
			SourceIPs: []string{"0.0.0.0/0", "::/0"},
			Port:      "any",
		},
	})
	updated.SetRName(res.RName())
	tmplMan := testtemplate.Manager{}
	resource.Test(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(firewall.ResourceType, firewall.ByID(t, &f)),
		Steps: []resource.TestStep{
			{
				// Create a new Firewall using the required values
				// only.
				Config: tmplMan.Render(t, "testdata/r/hcloud_firewall", res),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), firewall.ByID(t, &f)),
					resource.TestCheckResourceAttr(res.TFID(), "name",
						fmt.Sprintf("basic-firewall--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res.TFID(), "rule.#", "3"),
					testsupport.LiftTCF(hasFirewallRule(t, &f, "in", "80", "tcp", []string{"0.0.0.0/0", "::/0"}, []string{})),
					testsupport.LiftTCF(hasFirewallRule(t, &f, "in", "any", "udp", []string{"0.0.0.0/0", "::/0"}, []string{})),
					testsupport.LiftTCF(hasFirewallRule(t, &f, "out", "80", "tcp", []string{}, []string{"0.0.0.0/0", "::/0"})),
				),
			},
			{
				// Try to import the newly created Firewall
				ResourceName:      res.TFID(),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Create a new Firewall using the required values
				// only.
				Config: tmplMan.Render(t, "testdata/r/hcloud_firewall", updated),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), firewall.ByID(t, &f)),
					resource.TestCheckResourceAttr(res.TFID(), "name",
						fmt.Sprintf("basic-firewall--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res.TFID(), "rule.#", "3"),
					testsupport.LiftTCF(hasFirewallRule(t, &f, "in", "443", "tcp", []string{"0.0.0.0/0", "::/0"}, []string{})),
					testsupport.LiftTCF(hasFirewallRule(t, &f, "in", "any", "udp", []string{"0.0.0.0/0", "::/0"}, []string{})),
					testsupport.LiftTCF(hasFirewallRule(t, &f, "out", "443", "tcp", []string{}, []string{"0.0.0.0/0", "::/0"})),
				),
			},
		},
	})
}

func hasFirewallRule(
	t *testing.T,
	f *hcloud.Firewall,
	direction string,
	port string,
	protocol string, // nolint:unparam
	expectedSourceIps []string,
	expectedDestinationIps []string,
) func() error {
	return func() error {
		var firewallRule *hcloud.FirewallRule
		for _, r := range f.Rules {
			if string(r.Direction) == direction && *r.Port == port && string(r.Protocol) == protocol {
				firewallRule = &r
				break
			}
		}
		if !assert.NotNil(t, firewallRule, "firewall has no rule for this") {
			return nil
		}
		sourceIPs := make([]string, len(firewallRule.SourceIPs))
		for i, sourceIP := range firewallRule.SourceIPs {
			sourceIPs[i] = sourceIP.String()
		}

		destinationIPs := make([]string, len(firewallRule.DestinationIPs))
		for i, destinationIP := range firewallRule.DestinationIPs {
			destinationIPs[i] = destinationIP.String()
		}
		if assert.Exactly(t, expectedSourceIps, sourceIPs, "firewall rule does not contain expected source ips") {
			if assert.Exactly(t, expectedDestinationIps, destinationIPs, "firewall rule does not contain expected destination ips") {
				return nil
			}
		}
		return nil
	}
}
