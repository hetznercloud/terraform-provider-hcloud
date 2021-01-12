package rdns_test

import (
	"fmt"
	"testing"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/e2etests"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/sshkey"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/floatingip"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/rdns"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/server"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestRDNSResource_Server(t *testing.T) {
	var s hcloud.Server
	tmplMan := testtemplate.Manager{}

	sk := sshkey.NewRData(t, "server-rdns")
	resServer := &server.RData{
		Name:  "server-rdns",
		Type:  e2etests.TestServerType,
		Image: e2etests.TestImage,
		Labels: map[string]string{
			"tf-test": fmt.Sprintf("tf-test-rdns-%d", tmplMan.RandInt),
		},
		SSHKeys: []string{sk.TFID() + ".id"},
	}
	resServer.SetRName("server_rdns")
	resRDNS := rdns.NewRData(t, "rdnstest", resServer.TFID()+".id", "", resServer.TFID()+".ipv4_address", "example.hetzner.cloud")

	resource.Test(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, &s)),
		Steps: []resource.TestStep{
			{
				// Create a new RDNS using the required values
				// only.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", sk,
					"testdata/r/hcloud_server", resServer,
					"testdata/r/hcloud_rdns", resRDNS,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(resServer.TFID(), server.ByID(t, &s)),
					resource.TestCheckResourceAttr(resRDNS.TFID(), "dns_ptr", "example.hetzner.cloud"),
				),
			},
			{
				// Try to import the newly created RDNS
				ResourceName: resRDNS.TFID(),
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return fmt.Sprintf("s-%d-%s", s.ID, s.PublicNet.IPv4.IP.String()), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestRDNSResource_FloatingIP_IPv4(t *testing.T) {
	var fl hcloud.FloatingIP

	tmplMan := testtemplate.Manager{}
	restFloatingIP := &floatingip.RData{
		Name:             "floating-ipv4-rdns",
		Type:             "ipv4",
		HomeLocationName: "fsn1",
	}
	restFloatingIP.SetRName("floating_ips_rdns_v4")
	resRDNS := rdns.NewRData(t, "floating_ips_rdns_v4", "", restFloatingIP.TFID()+".id", restFloatingIP.TFID()+".ip_address", "example.hetzner.cloud")
	resource.Test(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(floatingip.ResourceType, floatingip.ByID(t, &fl)),
		Steps: []resource.TestStep{
			{
				// Create a new SSH Key using the required values
				// only.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_floating_ip", restFloatingIP,
					"testdata/r/hcloud_rdns", resRDNS,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(restFloatingIP.TFID(), floatingip.ByID(t, &fl)),
					resource.TestCheckResourceAttr("hcloud_rdns.floating_ips_rdns_v4", "dns_ptr", "example.hetzner.cloud"),
				),
			},
			{
				// Try to import the newly created RDNS
				ResourceName: resRDNS.TFID(),
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return fmt.Sprintf("f-%d-%s", fl.ID, fl.IP.String()), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestRDNSResource_FloatingIP_IPv6(t *testing.T) {
	var fl hcloud.FloatingIP

	tmplMan := testtemplate.Manager{}
	restFloatingIP := &floatingip.RData{
		Name:             "floating-ipv6-rdns",
		Type:             "ipv6",
		HomeLocationName: "fsn1",
	}
	restFloatingIP.SetRName("floating_ips_rdns_v6")

	resRDNS := rdns.NewRData(t, "floating_ips_rdns_v6", "", restFloatingIP.TFID()+".id", restFloatingIP.TFID()+".ip_address", "example.hetzner.cloud")
	resource.Test(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(floatingip.ResourceType, floatingip.ByID(t, &fl)),
		Steps: []resource.TestStep{
			{
				// Create a new SSH Key using the required values
				// only.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_floating_ip", restFloatingIP,
					"testdata/r/hcloud_rdns", resRDNS,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(restFloatingIP.TFID(), floatingip.ByID(t, &fl)),
					resource.TestCheckResourceAttr("hcloud_rdns.floating_ips_rdns_v6", "dns_ptr", "example.hetzner.cloud"),
				),
			},
			{
				// Try to import the newly created RDNS
				ResourceName: resRDNS.TFID(),
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return fmt.Sprintf("f-%d-%s", fl.ID, fl.IP.String()), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
