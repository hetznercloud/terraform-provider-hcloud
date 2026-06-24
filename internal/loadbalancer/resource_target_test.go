package loadbalancer_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/loadbalancer"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/network"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/server"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/sshkey"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testmux"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
)

func TestAccLoadBalancerTargetResource_ServerTarget(t *testing.T) {
	var (
		lb  hcloud.LoadBalancer
		srv hcloud.Server
	)

	tmplMan := testtemplate.Manager{}

	resSSHKey := sshkey.NewRData(t, "lb-server-target")
	resServer := &server.RData{
		Name:    "lb-server-target",
		Type:    teste2e.TestServerType,
		Image:   teste2e.TestImage,
		SSHKeys: []string{resSSHKey.TFID() + ".id"},
	}
	resServer.SetRName("lb-server-target")

	resLoadBalancer := &loadbalancer.RData{
		Name:        "target-test-lb",
		Type:        teste2e.TestLoadBalancerType,
		NetworkZone: "eu-central",
	}
	resLoadBalancer.SetRName(resLoadBalancer.Name)

	res1 := &loadbalancer.RDataTarget{
		Name:           "lb-test-target",
		Type:           "server",
		LoadBalancerID: resLoadBalancer.TFID() + ".id",
		ServerID:       resServer.TFID() + ".id",
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_server", resServer,
					"testdata/r/hcloud_load_balancer", resLoadBalancer,
					"testdata/r/hcloud_load_balancer_target", res1,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(resLoadBalancer.TFID(), loadbalancer.ByID(t, &lb)),
					testsupport.CheckResourceExists(resServer.TFID(), server.ByID(t, &srv)),
					resource.TestCheckResourceAttr(res1.TFID(), "type", "server"),
					testsupport.CheckResourceAttrFunc(res1.TFID(), "load_balancer_id", func() string {
						return util.FormatID(lb.ID)
					}),
					testsupport.CheckResourceAttrFunc(res1.TFID(), "server_id", func() string {
						return util.FormatID(srv.ID)
					}),
					testsupport.LiftTCF(hasServerTarget(&lb, &srv)),
				),
			},
			{
				ResourceName:      res1.TFID(),
				ImportState:       true,
				ImportStateIdFunc: loadBalancerTargetImportStateIDFunc("target-test-lb", hcloud.LoadBalancerTargetTypeServer, "lb-server-target"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLoadBalancerTargetResource_ServerTarget_UsePrivateIP(t *testing.T) {
	var (
		lb  hcloud.LoadBalancer
		srv hcloud.Server
	)

	tmplMan := testtemplate.Manager{}

	resSSHKey := sshkey.NewRData(t, "test")
	resServer := &server.RData{
		Name:    "lb-server-target-pi",
		Type:    teste2e.TestServerType,
		Image:   teste2e.TestImage,
		SSHKeys: []string{resSSHKey.TFID() + ".id"},
	}
	resServer.SetRName("test")

	resNetwork := &network.RData{
		Name:    "lb-target-test-network",
		IPRange: "10.0.0.0/16",
	}
	resNetwork.SetRName("test")

	resNetworkSubnet := &network.RDataSubnet{
		NetworkID:   resNetwork.TFID() + ".id",
		Type:        "cloud",
		NetworkZone: "eu-central",
		IPRange:     "10.0.1.0/24",
	}
	resNetworkSubnet.SetRName("test")

	resServerNetwork := &server.RDataNetwork{
		Name:      "lb-server-network",
		ServerID:  resServer.TFID() + ".id",
		NetworkID: resNetwork.TFID() + ".id",
	}
	resServerNetwork.SetRName("test")

	resLoadBalancer := &loadbalancer.RData{
		Name:        "target-test-lb",
		Type:        teste2e.TestLoadBalancerType,
		NetworkZone: "eu-central",
	}
	resLoadBalancer.SetRName("test")

	resLoadBalancerNetwork := &loadbalancer.RDataNetwork{
		Name:                  "target-test-lb-network",
		LoadBalancerID:        resLoadBalancer.TFID() + ".id",
		NetworkID:             resNetwork.TFID() + ".id",
		EnablePublicInterface: new(true),
	}
	resLoadBalancerNetwork.SetRName("test")

	res1 := &loadbalancer.RDataTarget{
		Name:           "target-test-lb",
		Type:           "server",
		LoadBalancerID: resLoadBalancer.TFID() + ".id",
		ServerID:       resServer.TFID() + ".id",
		UsePrivateIP:   true,
		DependsOn: []string{
			resServerNetwork.TFID(),
			resLoadBalancerNetwork.TFID(),
		},
	}
	res1.SetRName("target-test-lb")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_network", resNetwork,
					"testdata/r/hcloud_network_subnet", resNetworkSubnet,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_server", resServer,
					"testdata/r/hcloud_server_network", resServerNetwork,
					"testdata/r/hcloud_load_balancer", resLoadBalancer,
					"testdata/r/hcloud_load_balancer_network", resLoadBalancerNetwork,
					"testdata/r/hcloud_load_balancer_target", res1,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(resLoadBalancer.TFID(), loadbalancer.ByID(t, &lb)),
					testsupport.CheckResourceExists(resServer.TFID(), server.ByID(t, &srv)),
					resource.TestCheckResourceAttr(res1.TFID(), "use_private_ip", "true"),
					testsupport.LiftTCF(hasServerTarget(&lb, &srv)),
				),
			},
			{
				ResourceName:      res1.TFID(),
				ImportState:       true,
				ImportStateIdFunc: loadBalancerTargetImportStateIDFunc(resLoadBalancer.RName(), hcloud.LoadBalancerTargetTypeServer, resServer.RName()),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLoadBalancerTargetResource_LabelSelectorTarget(t *testing.T) {
	var (
		lb  hcloud.LoadBalancer
		srv hcloud.Server
	)

	tmplMan := testtemplate.Manager{}

	selector := fmt.Sprintf("tf-test=tf-test-%d", tmplMan.RandInt)

	resSSHKey := sshkey.NewRData(t, "lb-label-target")
	resServer := &server.RData{
		Name:  "lb-server-target",
		Type:  teste2e.TestServerType,
		Image: teste2e.TestImage,
		Labels: map[string]string{
			"tf-test": fmt.Sprintf("tf-test-%d", tmplMan.RandInt),
		},
		SSHKeys: []string{resSSHKey.TFID() + ".id"},
	}
	resServer.SetRName("lb-server-target")

	resLoadBalancer := &loadbalancer.RData{
		Name:        "target-test-lb",
		Type:        teste2e.TestLoadBalancerType,
		NetworkZone: "eu-central",
	}
	resLoadBalancer.SetRName("test")

	res1 := &loadbalancer.RDataTarget{
		Name:           "lb-test-target",
		Type:           "label_selector",
		LoadBalancerID: resLoadBalancer.TFID() + ".id",
		LabelSelector:  selector,
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_server", resServer,
					"testdata/r/hcloud_load_balancer", resLoadBalancer,
					"testdata/r/hcloud_load_balancer_target", res1,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(resLoadBalancer.TFID(), loadbalancer.ByID(t, &lb)),
					testsupport.CheckResourceExists(resServer.TFID(), server.ByID(t, &srv)),
					resource.TestCheckResourceAttr(res1.TFID(), "type", "label_selector"),
					resource.TestCheckResourceAttr(res1.TFID(), "label_selector", selector),
					testsupport.LiftTCF(hasLabelSelectorTarget(&lb, selector)),
				),
			},
			{
				ResourceName:      res1.TFID(),
				ImportState:       true,
				ImportStateIdFunc: loadBalancerTargetImportStateIDFunc(resLoadBalancer.RName(), hcloud.LoadBalancerTargetTypeLabelSelector, selector),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLoadBalancerTargetResource_LabelSelectorTarget_UsePrivateIP(t *testing.T) {
	var (
		lb  hcloud.LoadBalancer
		srv hcloud.Server
	)
	tmplMan := testtemplate.Manager{}

	resNetwork := &network.RData{
		Name:    "lb-target-test-network",
		IPRange: "10.0.0.0/16",
	}
	resNetwork.SetRName("test")

	resNetworkSubNet := &network.RDataSubnet{
		NetworkID:   resNetwork.TFID() + ".id",
		Type:        "cloud",
		NetworkZone: "eu-central",
		IPRange:     "10.0.1.0/24",
	}
	resNetworkSubNet.SetRName("test")

	resSSHKey := sshkey.NewRData(t, "test")
	resServer := &server.RData{
		Name:  "lb-server-target",
		Type:  teste2e.TestServerType,
		Image: teste2e.TestImage,
		Labels: map[string]string{
			"tf-test": fmt.Sprintf("tf-test-%d", tmplMan.RandInt),
		},
		SSHKeys: []string{resSSHKey.TFID() + ".id"},
	}
	resServer.SetRName("test")

	resServerNetwork := &server.RDataNetwork{
		Name:      "lb-server-network",
		ServerID:  resServer.TFID() + ".id",
		NetworkID: resNetwork.TFID() + ".id",
	}
	resServerNetwork.SetRName("test")

	resLoadBalancer := &loadbalancer.RData{
		Name:        "target-test-lb",
		Type:        teste2e.TestLoadBalancerType,
		NetworkZone: "eu-central",
	}
	resLoadBalancer.SetRName("test")

	resLoadBalancerNetwork := &loadbalancer.RDataNetwork{
		Name:                  "target-test-lb-network",
		LoadBalancerID:        resLoadBalancer.TFID() + ".id",
		NetworkID:             resNetwork.TFID() + ".id",
		EnablePublicInterface: new(true),
		DependsOn: []string{
			resNetworkSubNet.TFID(),
		},
	}
	resLoadBalancerNetwork.SetRName("test")

	selector := fmt.Sprintf("tf-test=tf-test-%d", tmplMan.RandInt)

	res1 := &loadbalancer.RDataTarget{
		Name:           "lb-test-target",
		Type:           "label_selector",
		LoadBalancerID: resLoadBalancer.TFID() + ".id",
		LabelSelector:  selector,
		UsePrivateIP:   true,
		DependsOn: []string{
			resServerNetwork.TFID(),
			resLoadBalancerNetwork.TFID(),
		},
	}
	res1.SetRName("test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_network", resNetwork,
					"testdata/r/hcloud_network_subnet", resNetworkSubNet,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_server", resServer,
					"testdata/r/hcloud_server_network", resServerNetwork,
					"testdata/r/hcloud_load_balancer", resLoadBalancer,
					"testdata/r/hcloud_load_balancer_network", resLoadBalancerNetwork,
					"testdata/r/hcloud_load_balancer_target", res1,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(resLoadBalancer.TFID(), loadbalancer.ByID(t, &lb)),
					testsupport.CheckResourceExists(resServer.TFID(), server.ByID(t, &srv)),
					resource.TestCheckResourceAttr(res1.TFID(), "type", "label_selector"),
					resource.TestCheckResourceAttr(res1.TFID(), "label_selector", selector),
					resource.TestCheckResourceAttr(res1.TFID(), "use_private_ip", "true"),
					testsupport.LiftTCF(hasLabelSelectorTarget(&lb, selector)),
				),
			},
			{
				ResourceName:      res1.TFID(),
				ImportState:       true,
				ImportStateIdFunc: loadBalancerTargetImportStateIDFunc(resLoadBalancer.RName(), hcloud.LoadBalancerTargetTypeLabelSelector, selector),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLoadBalancerTargetResource_IPTarget(t *testing.T) {
	t.Skip("No dedicated server available in test account")

	var (
		lb hcloud.LoadBalancer
	)

	ip := "213.239.214.25"

	resLoadBalancer := &loadbalancer.RData{
		Name:        "target-test-lb",
		Type:        teste2e.TestLoadBalancerType,
		NetworkZone: "eu-central",
	}
	resLoadBalancer.SetRName("test")

	res1 := &loadbalancer.RDataTarget{
		Name:           "lb-test-target",
		LoadBalancerID: resLoadBalancer.TFID() + ".id",
		Type:           "ip",
		IP:             ip,
	}
	res1.SetRName("test")

	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", resLoadBalancer,
					"testdata/r/hcloud_load_balancer_target", res1,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(resLoadBalancer.TFID(), loadbalancer.ByID(t, &lb)),
					resource.TestCheckResourceAttr(res1.TFID(), "type", "ip"),
					resource.TestCheckResourceAttr(res1.TFID(), "ip", ip),
					testsupport.LiftTCF(hasIPTarget(&lb, ip)),
				),
			},
			{
				ResourceName:      res1.TFID(),
				ImportState:       true,
				ImportStateIdFunc: loadBalancerTargetImportStateIDFunc(resLoadBalancer.RName(), hcloud.LoadBalancerTargetTypeIP, ip),
				ImportStateVerify: true,
			},
		},
	})
}

func hasServerTarget(lb *hcloud.LoadBalancer, srv *hcloud.Server) func() error {
	return func() error {
		for _, tgt := range lb.Targets {
			if tgt.Type == hcloud.LoadBalancerTargetTypeServer && tgt.Server.Server.ID == srv.ID {
				return nil
			}
		}
		return fmt.Errorf("load balancer %d: no target for server: %d", lb.ID, srv.ID)
	}
}

func hasLabelSelectorTarget(lb *hcloud.LoadBalancer, selector string) func() error {
	return func() error {
		for _, tgt := range lb.Targets {
			if tgt.Type == hcloud.LoadBalancerTargetTypeLabelSelector && tgt.LabelSelector.Selector == selector {
				return nil
			}
		}
		return fmt.Errorf("load balancer %d: no label selector: %s", lb.ID, selector)
	}
}

func hasIPTarget(lb *hcloud.LoadBalancer, ip string) func() error {
	return func() error {
		for _, tgt := range lb.Targets {
			if tgt.Type == hcloud.LoadBalancerTargetTypeIP && tgt.IP.IP == ip {
				return nil
			}
		}
		return fmt.Errorf("load balancer %d: no ip target: %s", lb.ID, ip)
	}
}

// loadBalancerTargetImportStateIDFunc builds the import ID for load balancer targets.
// In case the "server" type is used, pass the terraform resource name as the identifier, the
// function will automatically get the appropriate ID from terraform state.
// nolint:unparam
func loadBalancerTargetImportStateIDFunc(loadBalancerResourceName string, tgtType hcloud.LoadBalancerTargetType, identifier string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		lb, ok := s.RootModule().Resources[fmt.Sprintf("%s.%s", loadbalancer.ResourceType, loadBalancerResourceName)]
		if !ok {
			return "", fmt.Errorf("load balancer not found: %s", loadBalancerResourceName)
		}

		if tgtType == hcloud.LoadBalancerTargetTypeServer {
			server, ok := s.RootModule().Resources[fmt.Sprintf("%s.%s", server.ResourceType, identifier)]
			if !ok {
				return "", fmt.Errorf("server not found: %s", identifier)
			}
			identifier = server.Primary.ID
		}

		return fmt.Sprintf("%s__%s__%s", lb.Primary.ID, tgtType, identifier), nil
	}
}
