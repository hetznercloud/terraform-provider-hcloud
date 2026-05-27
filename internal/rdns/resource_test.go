package rdns_test

import (
	"fmt"
	"net"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/kit/randutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/floatingip"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/loadbalancer"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/primaryip"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/rdns"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/server"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/sshkey"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccRDNSResource_Errors(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_rdns", rdns.NewRDataPrimaryIP(t,
						"ipv6",
						"12345", // Not found
						fmt.Sprintf("%q", "2001:0db8::231"),
						"ipv6.example.org",
					),
				),
				ExpectError: regexp.MustCompile(`Resource \(primary ip\) was not found: id=12345`),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_rdns", rdns.NewRDataPrimaryIP(t,
						"ipv6",
						"12345",
						fmt.Sprintf("%q", "2001"), // Invalid ip address
						"ipv6.example.org",
					),
				),
				ExpectError: regexp.MustCompile(`Invalid IP Address String Value`),
			},
		},
	})
}

func TestAccRDNSResource_Server(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	var hcServer hcloud.Server

	resSSHKey := sshkey.NewRData(t, "main")
	resServer := &server.RData{
		Name:    randutil.GenerateID(),
		Type:    teste2e.TestServerType,
		Image:   teste2e.TestImage,
		SSHKeys: []string{resSSHKey.TFID() + ".id"},
	}
	resServer.SetRName("main")

	// IPv4
	resA1 := rdns.NewRDataServer(t,
		"ipv4",
		resServer.TFID()+".id",
		resServer.TFID()+".ipv4_address",
		"ipv4.example.org",
	)
	resA2 := testtemplate.DeepCopy(t, resA1)
	resA2.DNSPTR = "changed." + resA1.DNSPTR

	// IPv6
	resB1 := rdns.NewRDataServer(t,
		"ipv6",
		resServer.TFID()+".id",
		resServer.TFID()+".ipv6_address",
		"ipv6.example.org",
	)
	resB2 := testtemplate.DeepCopy(t, resB1)
	resB2.DNSPTR = "changed." + resB1.DNSPTR

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(server.ResourceType, server.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_server", resServer,
					"testdata/r/hcloud_rdns", resA1,
					"testdata/r/hcloud_rdns", resB1,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(resServer.TFID(), server.ByID(t, &hcServer)),
					resource.TestCheckResourceAttr(resA1.TFID(), "dns_ptr", resA1.DNSPTR),
					resource.TestCheckResourceAttr(resB1.TFID(), "dns_ptr", resB1.DNSPTR),
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_server", resServer,
					"testdata/r/hcloud_rdns", resA2,
					"testdata/r/hcloud_rdns", resB2,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(resServer.TFID(), server.ByID(t, &hcServer)),
					resource.TestCheckResourceAttr(resA2.TFID(), "dns_ptr", resA2.DNSPTR),
					resource.TestCheckResourceAttr(resB2.TFID(), "dns_ptr", resB2.DNSPTR),
				),
			},
			{
				ResourceName: resA2.TFID(),
				ImportStateIdFunc: func(_ *terraform.State) (string, error) {
					return rdns.FormatID(&hcServer, hcServer.PublicNet.IPv4.IP), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName: resB2.TFID(),
				ImportStateIdFunc: func(_ *terraform.State) (string, error) {
					ip := net.ParseIP(hcServer.PublicNet.IPv6.IP.String() + "1")
					return rdns.FormatID(&hcServer, ip), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRDNSResource_Server_UpgradePluginFramework(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	resSSHKey := sshkey.NewRData(t, "main")
	resServer := &server.RData{
		Name:    randutil.GenerateID(),
		Type:    teste2e.TestServerType,
		Image:   teste2e.TestImage,
		SSHKeys: []string{resSSHKey.TFID() + ".id"},
	}
	resServer.SetRName("main")

	res := rdns.NewRDataServer(t,
		"ipv4",
		resServer.TFID()+".id",
		resServer.TFID()+".ipv4_address",
		"ipv4.example.org",
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: teste2e.PreCheck(t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"hcloud": {
						VersionConstraint: "1.63.0",
						Source:            "hetznercloud/hcloud",
					},
				},
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_server", resServer,
					"testdata/r/hcloud_rdns", res,
				),
			},
			{
				ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_server", resServer,
					"testdata/r/hcloud_rdns", res,
				),
				PlanOnly: true,
			},
		},
	})
}

func TestAccRDNSResource_PrimaryIP(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	var (
		hcPrimaryIPv4 hcloud.PrimaryIP
		hcPrimaryIPv6 hcloud.PrimaryIP
	)

	resPrimaryIPv4 := &primaryip.RData{
		Name:     randutil.GenerateID(),
		Type:     "ipv4",
		Location: teste2e.TestLocationName,
	}
	resPrimaryIPv4.SetRName("ipv4")

	resPrimaryIPv6 := &primaryip.RData{
		Name:     randutil.GenerateID(),
		Type:     "ipv6",
		Location: teste2e.TestLocationName,
	}
	resPrimaryIPv6.SetRName("ipv6")

	// IPv4
	resA1 := rdns.NewRDataPrimaryIP(t,
		"ipv4",
		resPrimaryIPv4.TFID()+".id",
		resPrimaryIPv4.TFID()+".ip_address",
		"ipv4.example.org",
	)

	resA2 := testtemplate.DeepCopy(t, resA1)
	resA2.DNSPTR = "changed." + resA1.DNSPTR

	// IPv6
	resB1 := rdns.NewRDataPrimaryIP(t,
		"ipv6",
		resPrimaryIPv6.TFID()+".id",
		resPrimaryIPv6.TFID()+".ip_address",
		"ipv6.example.org",
	)

	resB2 := testtemplate.DeepCopy(t, resB1)
	resB2.DNSPTR = "changed." + resB1.DNSPTR

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(primaryip.ResourceType, primaryip.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_primary_ip", resPrimaryIPv4,
					"testdata/r/hcloud_primary_ip", resPrimaryIPv6,
					"testdata/r/hcloud_rdns", resA1,
					"testdata/r/hcloud_rdns", resB1,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(resPrimaryIPv4.TFID(), primaryip.ByID(t, &hcPrimaryIPv4)),
					testsupport.CheckResourceExists(resPrimaryIPv6.TFID(), primaryip.ByID(t, &hcPrimaryIPv6)),
					resource.TestCheckResourceAttr(resA1.TFID(), "dns_ptr", resA1.DNSPTR),
					resource.TestCheckResourceAttr(resB1.TFID(), "dns_ptr", resB1.DNSPTR),
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_primary_ip", resPrimaryIPv4,
					"testdata/r/hcloud_primary_ip", resPrimaryIPv6,
					"testdata/r/hcloud_rdns", resA2,
					"testdata/r/hcloud_rdns", resB2,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(resPrimaryIPv4.TFID(), primaryip.ByID(t, &hcPrimaryIPv4)),
					testsupport.CheckResourceExists(resPrimaryIPv6.TFID(), primaryip.ByID(t, &hcPrimaryIPv6)),
					resource.TestCheckResourceAttr(resA2.TFID(), "dns_ptr", resA2.DNSPTR),
					resource.TestCheckResourceAttr(resB2.TFID(), "dns_ptr", resB2.DNSPTR),
				),
			},
			{
				ResourceName: resA2.TFID(),
				ImportStateIdFunc: func(_ *terraform.State) (string, error) {
					return rdns.FormatID(&hcPrimaryIPv4, hcPrimaryIPv4.IP), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName: resB2.TFID(),
				ImportStateIdFunc: func(_ *terraform.State) (string, error) {
					return rdns.FormatID(&hcPrimaryIPv6, hcPrimaryIPv6.IP), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRDNSResource_PrimaryIP_UpgradePluginFramework(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	resPrimaryIP := &primaryip.RData{
		Name:     randutil.GenerateID(),
		Type:     "ipv6",
		Location: teste2e.TestLocationName,
	}
	resPrimaryIP.SetRName("ipv6")

	res := rdns.NewRDataPrimaryIP(t,
		"ipv6",
		resPrimaryIP.TFID()+".id",
		resPrimaryIP.TFID()+".ip_address",
		"ipv6.example.org",
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: teste2e.PreCheck(t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"hcloud": {
						VersionConstraint: "1.63.0",
						Source:            "hetznercloud/hcloud",
					},
				},
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_primary_ip", resPrimaryIP,
					"testdata/r/hcloud_rdns", res,
				),
			},
			{
				ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_primary_ip", resPrimaryIP,
					"testdata/r/hcloud_rdns", res,
				),
				PlanOnly: true,
			},
		},
	})
}

func TestAccRDNSResource_FloatingIP(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	var (
		hcFloatingIPv4 hcloud.FloatingIP
		hcFloatingIPv6 hcloud.FloatingIP
	)

	resFloatingIPv4 := &floatingip.RData{
		Name:             randutil.GenerateID(),
		Type:             "ipv4",
		HomeLocationName: teste2e.TestLocationName,
	}
	resFloatingIPv4.SetRName("ipv4")

	resFloatingIPv6 := &floatingip.RData{
		Name:             randutil.GenerateID(),
		Type:             "ipv6",
		HomeLocationName: teste2e.TestLocationName,
	}
	resFloatingIPv6.SetRName("ipv6")

	// IPv4
	resA1 := rdns.NewRDataFloatingIP(t,
		"ipv4",
		resFloatingIPv4.TFID()+".id",
		resFloatingIPv4.TFID()+".ip_address",
		"ipv4.example.org",
	)

	resA2 := testtemplate.DeepCopy(t, resA1)
	resA2.DNSPTR = "changed." + resA1.DNSPTR

	// IPv6
	resB1 := rdns.NewRDataFloatingIP(t,
		"ipv6",
		resFloatingIPv6.TFID()+".id",
		resFloatingIPv6.TFID()+".ip_address",
		"ipv6.example.org",
	)

	resB2 := testtemplate.DeepCopy(t, resB1)
	resB2.DNSPTR = "changed." + resB1.DNSPTR

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(floatingip.ResourceType, floatingip.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_floating_ip", resFloatingIPv4,
					"testdata/r/hcloud_floating_ip", resFloatingIPv6,
					"testdata/r/hcloud_rdns", resA1,
					"testdata/r/hcloud_rdns", resB1,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(resFloatingIPv4.TFID(), floatingip.ByID(t, &hcFloatingIPv4)),
					testsupport.CheckResourceExists(resFloatingIPv6.TFID(), floatingip.ByID(t, &hcFloatingIPv6)),
					resource.TestCheckResourceAttr(resA1.TFID(), "dns_ptr", resA1.DNSPTR),
					resource.TestCheckResourceAttr(resB1.TFID(), "dns_ptr", resB1.DNSPTR),
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_floating_ip", resFloatingIPv4,
					"testdata/r/hcloud_floating_ip", resFloatingIPv6,
					"testdata/r/hcloud_rdns", resA2,
					"testdata/r/hcloud_rdns", resB2,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(resFloatingIPv4.TFID(), floatingip.ByID(t, &hcFloatingIPv4)),
					testsupport.CheckResourceExists(resFloatingIPv6.TFID(), floatingip.ByID(t, &hcFloatingIPv6)),
					resource.TestCheckResourceAttr(resA2.TFID(), "dns_ptr", resA2.DNSPTR),
					resource.TestCheckResourceAttr(resB2.TFID(), "dns_ptr", resB2.DNSPTR),
				),
			},
			{
				ResourceName: resA2.TFID(),
				ImportStateIdFunc: func(_ *terraform.State) (string, error) {
					return rdns.FormatID(&hcFloatingIPv4, hcFloatingIPv4.IP), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName: resB2.TFID(),
				ImportStateIdFunc: func(_ *terraform.State) (string, error) {
					return rdns.FormatID(&hcFloatingIPv6, hcFloatingIPv6.IP), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRDNSResource_FloatingIP_UpgradePluginFramework(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	resFloatingIP := &floatingip.RData{
		Name:             randutil.GenerateID(),
		Type:             "ipv4",
		HomeLocationName: teste2e.TestLocationName,
	}
	resFloatingIP.SetRName("ipv4")

	res := rdns.NewRDataFloatingIP(t,
		"ipv4",
		resFloatingIP.TFID()+".id",
		resFloatingIP.TFID()+".ip_address",
		"ipv4.example.org",
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: teste2e.PreCheck(t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"hcloud": {
						VersionConstraint: "1.63.0",
						Source:            "hetznercloud/hcloud",
					},
				},
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_floating_ip", resFloatingIP,
					"testdata/r/hcloud_rdns", res,
				),
			},
			{
				ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_floating_ip", resFloatingIP,
					"testdata/r/hcloud_rdns", res,
				),
				PlanOnly: true,
			},
		},
	})
}

func TestAccRDNSResource_LoadBalancer(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	var hcLoadBalancer hcloud.LoadBalancer

	resLoadBalancer := &loadbalancer.RData{
		Name:         randutil.GenerateID(),
		LocationName: teste2e.TestLocationName,
	}
	resLoadBalancer.SetRName("main")

	// IPv4
	resA1 := rdns.NewRDataLoadBalancer(t,
		"ipv4",
		resLoadBalancer.TFID()+".id",
		resLoadBalancer.TFID()+".ipv4",
		"ipv4.example.org",
	)

	resA2 := testtemplate.DeepCopy(t, resA1)
	resA2.DNSPTR = "changed." + resA1.DNSPTR

	// IPv6
	resB1 := rdns.NewRDataLoadBalancer(t,
		"ipv6",
		resLoadBalancer.TFID()+".id",
		resLoadBalancer.TFID()+".ipv6",
		"ipv6.example.org",
	)

	resB2 := testtemplate.DeepCopy(t, resB1)
	resB2.DNSPTR = "changed." + resB1.DNSPTR

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),

		CheckDestroy: testsupport.CheckAPIResourceAllAbsent(loadbalancer.ResourceType, loadbalancer.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", resLoadBalancer,
					"testdata/r/hcloud_rdns", resA1,
					"testdata/r/hcloud_rdns", resB1,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(resLoadBalancer.TFID(), loadbalancer.ByID(t, &hcLoadBalancer)),
					resource.TestCheckResourceAttr(resA1.TFID(), "dns_ptr", resA1.DNSPTR),
					resource.TestCheckResourceAttr(resB1.TFID(), "dns_ptr", resB1.DNSPTR),
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", resLoadBalancer,
					"testdata/r/hcloud_rdns", resA2,
					"testdata/r/hcloud_rdns", resB2,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(resLoadBalancer.TFID(), loadbalancer.ByID(t, &hcLoadBalancer)),
					resource.TestCheckResourceAttr(resA2.TFID(), "dns_ptr", resA2.DNSPTR),
					resource.TestCheckResourceAttr(resB2.TFID(), "dns_ptr", resB2.DNSPTR),
				),
			},
			{
				ResourceName: resA2.TFID(),
				ImportStateIdFunc: func(_ *terraform.State) (string, error) {
					return rdns.FormatID(&hcLoadBalancer, hcLoadBalancer.PublicNet.IPv4.IP), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName: resB2.TFID(),
				ImportStateIdFunc: func(_ *terraform.State) (string, error) {
					return rdns.FormatID(&hcLoadBalancer, hcLoadBalancer.PublicNet.IPv6.IP), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRDNSResource_LoadBalancer_UpgradePluginFramework(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	resLoadBalancer := &loadbalancer.RData{
		Name:         randutil.GenerateID(),
		LocationName: teste2e.TestLocationName,
	}
	resLoadBalancer.SetRName("main")

	res := rdns.NewRDataLoadBalancer(t,
		"ipv4",
		resLoadBalancer.TFID()+".id",
		resLoadBalancer.TFID()+".ipv4",
		"ipv4.example.org",
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: teste2e.PreCheck(t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"hcloud": {
						VersionConstraint: "1.63.0",
						Source:            "hetznercloud/hcloud",
					},
				},
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", resLoadBalancer,
					"testdata/r/hcloud_rdns", res,
				),
			},
			{
				ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", resLoadBalancer,
					"testdata/r/hcloud_rdns", res,
				),
				PlanOnly: true,
			},
		},
	})
}
