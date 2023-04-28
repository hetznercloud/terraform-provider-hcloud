package rdns_test

import (
	"fmt"

	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/e2etests"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/floatingip"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/loadbalancer"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/primaryip"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/rdns"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/server"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/sshkey"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestRDNSResource_Server(t *testing.T) {
	tests := []struct {
		name        string
		dns         string
		ipAddress   string
		ipAsStrFunc func(s *hcloud.Server) string
	}{
		{
			name:      "server-ipv6-rdns",
			dns:       "ipv6.example.hetzner.cloud",
			ipAddress: ".ipv6_address",
			ipAsStrFunc: func(s *hcloud.Server) string {
				return s.PublicNet.IPv6.IP.String() + "1"
			},
		},
		{
			name:      "server-ipv4-rdns",
			dns:       "ipv4.example.hetzner.cloud",
			ipAddress: ".ipv4_address",
			ipAsStrFunc: func(s *hcloud.Server) string {
				return s.PublicNet.IPv4.IP.String()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s hcloud.Server
			tmplMan := testtemplate.Manager{}

			sk := sshkey.NewRData(t, tt.name)
			resServer := &server.RData{
				Name:  tt.name,
				Type:  e2etests.TestServerType,
				Image: e2etests.TestImage,
				Labels: map[string]string{
					"tf-test": fmt.Sprintf("tf-test-rdns-%d", tmplMan.RandInt),
				},
				SSHKeys: []string{sk.TFID() + ".id"},
			}
			resServer.SetRName(tt.name)
			resRDNS := rdns.NewRDataServer(t, tt.name, resServer.TFID()+".id", resServer.TFID()+tt.ipAddress, tt.dns)

			// TODO: Debug issues that causes this to fail when running in parallel
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
							resource.TestCheckResourceAttr(resRDNS.TFID(), "dns_ptr", tt.dns),
						),
					},
					{
						// Try to import the newly created RDNS
						ResourceName: resRDNS.TFID(),
						ImportStateIdFunc: func(state *terraform.State) (string, error) {
							return fmt.Sprintf("s-%d-%s", s.ID, tt.ipAsStrFunc(&s)), nil
						},
						ImportState:       true,
						ImportStateVerify: true,
					},
				},
			})
		})
	}
}

func TestRDNSResource_PrimaryIP(t *testing.T) {
	tests := []struct {
		name          string
		dns           string
		primaryIPType string
	}{
		{
			name:          "primary-ipv6-rdns",
			dns:           "ipv6.example.hetzner.cloud",
			primaryIPType: "ipv6",
		},
		{
			name:          "primary-ipv4-rdns",
			dns:           "ipv4.example.hetzner.cloud",
			primaryIPType: "ipv4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var primaryIP hcloud.PrimaryIP

			tmplMan := testtemplate.Manager{}
			restPrimaryIP := &primaryip.RData{
				Name:         tt.name,
				Type:         tt.primaryIPType,
				AssigneeType: "server",
				Datacenter:   e2etests.TestDataCenter,
			}
			restPrimaryIP.SetRName(tt.name)
			resRDNS := rdns.NewRDataPrimaryIP(t, tt.name, restPrimaryIP.TFID()+".id", restPrimaryIP.TFID()+".ip_address", tt.dns)

			// TODO: Debug issues that causes this to fail when running in parallel
			resource.Test(t, resource.TestCase{
				PreCheck:     e2etests.PreCheck(t),
				Providers:    e2etests.Providers(),
				CheckDestroy: testsupport.CheckResourcesDestroyed(primaryip.ResourceType, primaryip.ByID(t, &primaryIP)),
				Steps: []resource.TestStep{
					{
						// Create a new SSH Key using the required values
						// only.
						Config: tmplMan.Render(t,
							"testdata/r/hcloud_primary_ip", restPrimaryIP,
							"testdata/r/hcloud_rdns", resRDNS,
						),
						Check: resource.ComposeTestCheckFunc(
							testsupport.CheckResourceExists(restPrimaryIP.TFID(), primaryip.ByID(t, &primaryIP)),
							resource.TestCheckResourceAttr(resRDNS.TFID(), "dns_ptr", tt.dns),
						),
					},
					{
						// Try to import the newly created RDNS
						ResourceName: resRDNS.TFID(),
						ImportStateIdFunc: func(state *terraform.State) (string, error) {
							return fmt.Sprintf("p-%d-%s", primaryIP.ID, primaryIP.IP.String()), nil
						},
						ImportState:       true,
						ImportStateVerify: true,
					},
				},
			})
		})
	}
}

func TestRDNSResource_FloatingIP(t *testing.T) {
	tests := []struct {
		name           string
		dns            string
		floatingIPType string
	}{
		{
			name:           "floating-ipv6-rdns",
			dns:            "ipv6.example.hetzner.cloud",
			floatingIPType: "ipv6",
		},
		{
			name:           "floating-ipv4-rdns",
			dns:            "ipv4.example.hetzner.cloud",
			floatingIPType: "ipv4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fl hcloud.FloatingIP

			tmplMan := testtemplate.Manager{}
			restFloatingIP := &floatingip.RData{
				Name:             tt.name,
				Type:             tt.floatingIPType,
				HomeLocationName: e2etests.TestLocationName,
			}
			restFloatingIP.SetRName(tt.name)
			resRDNS := rdns.NewRDataFloatingIP(t, tt.name, restFloatingIP.TFID()+".id", restFloatingIP.TFID()+".ip_address", tt.dns)

			// TODO: Debug issues that causes this to fail when running in parallel
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
							resource.TestCheckResourceAttr(resRDNS.TFID(), "dns_ptr", tt.dns),
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
		})
	}
}

func TestRDNSResource_LoadBalancer(t *testing.T) {
	tests := []struct {
		name        string
		ipAddress   string
		dns         string
		ipAsStrFunc func(lb *hcloud.LoadBalancer) string
	}{
		{
			name:      "load-balancer-ipv6-rdns",
			ipAddress: ".ipv6",
			dns:       "ipv6.example.hetzner.cloud",
			ipAsStrFunc: func(lb *hcloud.LoadBalancer) string {
				return lb.PublicNet.IPv6.IP.String()
			},
		},
		{
			name:      "load-balancer-ipv4-rdns",
			ipAddress: ".ipv4",
			dns:       "ipv4.example.hetzner.cloud",
			ipAsStrFunc: func(lb *hcloud.LoadBalancer) string {
				return lb.PublicNet.IPv4.IP.String()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var lb hcloud.LoadBalancer

			tmplMan := testtemplate.Manager{}
			restLoadBalancer := &loadbalancer.RData{
				Name:         tt.name,
				LocationName: e2etests.TestLocationName,
			}
			restLoadBalancer.SetRName(tt.name)

			resRDNS := rdns.NewRDataLoadBalancer(t, tt.name, restLoadBalancer.TFID()+".id", restLoadBalancer.TFID()+tt.ipAddress, tt.dns)

			// TODO: Debug issues that causes this to fail when running in parallel
			resource.Test(t, resource.TestCase{
				PreCheck:     e2etests.PreCheck(t),
				Providers:    e2etests.Providers(),
				CheckDestroy: testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, &lb)),
				Steps: []resource.TestStep{
					{
						Config: tmplMan.Render(t,
							"testdata/r/hcloud_load_balancer", restLoadBalancer,
							"testdata/r/hcloud_rdns", resRDNS,
						),
						Check: resource.ComposeTestCheckFunc(
							testsupport.CheckResourceExists(restLoadBalancer.TFID(), loadbalancer.ByID(t, &lb)),
							resource.TestCheckResourceAttr(resRDNS.TFID(), "dns_ptr", tt.dns),
						),
					},
					{
						// Try to import the newly created RDNS
						ResourceName: resRDNS.TFID(),
						ImportStateIdFunc: func(state *terraform.State) (string, error) {
							return fmt.Sprintf("l-%d-%s", lb.ID, tt.ipAsStrFunc(&lb)), nil
						},
						ImportState:       true,
						ImportStateVerify: true,
					},
				},
			})
		})
	}
}
