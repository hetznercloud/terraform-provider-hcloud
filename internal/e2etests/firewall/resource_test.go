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
			Direction:   "in",
			Protocol:    "tcp",
			SourceIPs:   []string{"0.0.0.0/0", "::/0"},
			Port:        "80",
			Description: "allow http in",
		},
		{
			Direction:      "out",
			Protocol:       "tcp",
			DestinationIPs: []string{"0.0.0.0/0", "::/0"},
			Port:           "80",
			Description:    "allow http out",
		},
		{
			Direction:   "in",
			Protocol:    "udp",
			SourceIPs:   []string{"0.0.0.0/0", "::/0"},
			Port:        "any",
			Description: "allow udp in all ports",
		},
	}, nil)

	updated := firewall.NewRData(t, "basic-firewall", []firewall.RDataRule{
		{
			Direction:   "in",
			Protocol:    "tcp",
			SourceIPs:   []string{"0.0.0.0/0", "::/0"},
			Port:        "443",
			Description: "allow https in",
		},
		{
			Direction:      "out",
			Protocol:       "tcp",
			DestinationIPs: []string{"0.0.0.0/0", "::/0"},
			Port:           "443",
			Description:    "allow https out",
		},
		{
			Direction:   "in",
			Protocol:    "udp",
			SourceIPs:   []string{"0.0.0.0/0", "::/0"},
			Port:        "any",
			Description: "allow udp in all ports",
		},
	}, nil)
	updated.SetRName(res.RName())
	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
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
					testsupport.LiftTCF(hasFirewallRule(t, &f, "in", "80", "tcp", []string{"0.0.0.0/0", "::/0"}, []string{}, "allow http in")),
					testsupport.LiftTCF(hasFirewallRule(t, &f, "in", "any", "udp", []string{"0.0.0.0/0", "::/0"}, []string{}, "allow udp in all ports")),
					testsupport.LiftTCF(hasFirewallRule(t, &f, "out", "80", "tcp", []string{}, []string{"0.0.0.0/0", "::/0"}, "allow http out")),
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
					testsupport.LiftTCF(hasFirewallRule(t, &f, "in", "443", "tcp", []string{"0.0.0.0/0", "::/0"}, []string{}, "allow https in")),
					testsupport.LiftTCF(hasFirewallRule(t, &f, "in", "any", "udp", []string{"0.0.0.0/0", "::/0"}, []string{}, "allow udp in all ports")),
					testsupport.LiftTCF(hasFirewallRule(t, &f, "out", "443", "tcp", []string{}, []string{"0.0.0.0/0", "::/0"}, "allow https out")),
				),
			},
		},
	})
}

func TestFirewallResource_ApplyTo(t *testing.T) {
	var f hcloud.Firewall

	res := firewall.NewRData(t, "applyto-firewall", []firewall.RDataRule{
		{
			Direction:   "in",
			Protocol:    "tcp",
			SourceIPs:   []string{"0.0.0.0/0", "::/0"},
			Port:        "80",
			Description: "allow http in",
		},
	}, []firewall.RDataApplyTo{
		{
			LabelSelector: "key=value",
		},
	})
	resUpdated := firewall.NewRData(t, "applyto-firewall", []firewall.RDataRule{
		{
			Direction:   "in",
			Protocol:    "tcp",
			SourceIPs:   []string{"0.0.0.0/0", "::/0"},
			Port:        "80",
			Description: "allow http in",
		},
	}, []firewall.RDataApplyTo{
		{
			LabelSelector: "key=value",
		},
		{
			LabelSelector: "another=value",
		},
	})
	resUpdated.SetRName(res.RName())
	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
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
						fmt.Sprintf("applyto-firewall--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res.TFID(), "rule.#", "1"),
					testsupport.LiftTCF(hasFirewallRule(t, &f, "in", "80", "tcp", []string{"0.0.0.0/0", "::/0"}, []string{}, "allow http in")),
					testsupport.LiftTCF(hasLabelSelectorResource(t, &f, "key=value")),
				),
			},
			{
				Config: tmplMan.Render(t, "testdata/r/hcloud_firewall", resUpdated),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(resUpdated.TFID(), firewall.ByID(t, &f)),
					resource.TestCheckResourceAttr(resUpdated.TFID(), "name",
						fmt.Sprintf("applyto-firewall--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(resUpdated.TFID(), "rule.#", "1"),
					testsupport.LiftTCF(hasFirewallRule(t, &f, "in", "80", "tcp", []string{"0.0.0.0/0", "::/0"}, []string{}, "allow http in")),
					testsupport.LiftTCF(hasLabelSelectorResource(t, &f, "key=value")),
					testsupport.LiftTCF(hasLabelSelectorResource(t, &f, "another=value")),
				),
			},
		},
	})
}

func TestFirewallResource_SourceIPs_IPv6Comparison(t *testing.T) {
	var f hcloud.Firewall

	res := firewall.NewRData(t, "ipv6-firewall", []firewall.RDataRule{
		{
			Direction: "in",
			Protocol:  "tcp",
			SourceIPs: []string{"Aaaa:aaaa:aaaa:aaaa::/64"},
			Port:      "22",
		},
	}, nil)
	tmplMan := testtemplate.Manager{}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(firewall.ResourceType, firewall.ByID(t, &f)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t, "testdata/r/hcloud_firewall", res),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), firewall.ByID(t, &f)),
				),
			},
			{
				Config:   tmplMan.Render(t, "testdata/r/hcloud_firewall", res),
				PlanOnly: true,
			},
		},
	})
}

func TestFirewallResource_DestinationIPs_IPv6Comparison(t *testing.T) {
	var f hcloud.Firewall

	res := firewall.NewRData(t, "ipv6-firewall", []firewall.RDataRule{
		{
			Direction:      "out",
			Protocol:       "tcp",
			DestinationIPs: []string{"Aaaa:aaaa:aaaa:aaaa::/64"},
			Port:           "22",
		},
	}, nil)
	tmplMan := testtemplate.Manager{}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(firewall.ResourceType, firewall.ByID(t, &f)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t, "testdata/r/hcloud_firewall", res),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), firewall.ByID(t, &f)),
				),
			},
			{
				Config:   tmplMan.Render(t, "testdata/r/hcloud_firewall", res),
				PlanOnly: true,
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
	description string,
) func() error {
	return func() error {
		var firewallRule *hcloud.FirewallRule
		for _, r := range f.Rules {
			if string(r.Direction) == direction && *r.Port == port && string(r.Protocol) == protocol && *r.Description == description {
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

func hasLabelSelectorResource(t *testing.T, f *hcloud.Firewall, labelSelector string) func() error {
	return func() error {
		var firewallResource *hcloud.FirewallResource

		for _, r := range f.AppliedTo {
			if r.Type == hcloud.FirewallResourceTypeLabelSelector && r.LabelSelector.Selector == labelSelector {
				firewallResource = &r
				break
			}
		}
		if !assert.NotNil(t, firewallResource, "firewall has no resource for this") {
			return nil
		}
		return nil
	}
}

func hasServerResource(t *testing.T, fw *hcloud.Firewall, srv *hcloud.Server) func() error {
	return func() error {
		for _, r := range fw.AppliedTo {
			if r.Type == hcloud.FirewallResourceTypeServer && r.Server.ID == srv.ID {
				return nil
			}
		}

		t.Errorf("Firewall %d has no server resource for server %d", fw.ID, srv.ID)
		return nil
	}
}
