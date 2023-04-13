package loadbalancer_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/e2etests"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/sshkey"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/loadbalancer"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/network"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/server"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccHcloudLoadBalancerTarget_ServerTarget(t *testing.T) {
	var (
		lb  hcloud.LoadBalancer
		srv hcloud.Server
	)

	tmplMan := testtemplate.Manager{}

	resSSHKey := sshkey.NewRData(t, "lb-server-target")
	resServer := &server.RData{
		Name:    "lb-server-target",
		Type:    e2etests.TestServerType,
		Image:   e2etests.TestImage,
		SSHKeys: []string{resSSHKey.TFID() + ".id"},
	}
	resServer.SetRName("lb-server-target")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_server", resServer,
					"testdata/r/hcloud_load_balancer", &loadbalancer.RData{
						Name:        "target-test-lb",
						Type:        e2etests.TestLoadBalancerType,
						NetworkZone: "eu-central",
					},
					"testdata/r/hcloud_load_balancer_target", &loadbalancer.RDataTarget{
						Name:           "lb-test-target",
						Type:           "server",
						LoadBalancerID: "hcloud_load_balancer.target-test-lb.id",
						ServerID:       resServer.TFID() + ".id",
					},
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(
						loadbalancer.ResourceType+".target-test-lb", loadbalancer.ByID(t, &lb)),
					testsupport.CheckResourceExists(resServer.TFID(), server.ByID(t, &srv)),
					resource.TestCheckResourceAttr(
						loadbalancer.TargetResourceType+".lb-test-target", "type", "server"),
					testsupport.CheckResourceAttrFunc(
						loadbalancer.TargetResourceType+".lb-test-target", "load_balancer_id", func() string {
							return strconv.Itoa(lb.ID)
						}),
					testsupport.CheckResourceAttrFunc(
						loadbalancer.TargetResourceType+".lb-test-target", "server_id", func() string {
							return strconv.Itoa(srv.ID)
						}),
					testsupport.LiftTCF(hasServerTarget(&lb, &srv)),
				),
			},
			{
				ResourceName:      loadbalancer.TargetResourceType + ".lb-test-target",
				ImportState:       true,
				ImportStateIdFunc: loadBalancerTargetImportStateIDFunc("target-test-lb", hcloud.LoadBalancerTargetTypeServer, "lb-server-target"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccHcloudLoadBalancerTarget_ServerTarget_UsePrivateIP(t *testing.T) {
	var (
		lb  hcloud.LoadBalancer
		srv hcloud.Server
	)

	tmplMan := testtemplate.Manager{}

	resSSHKey := sshkey.NewRData(t, "lb-server-target-pi")
	resServer := &server.RData{
		Name:    "lb-server-target-pi",
		Type:    e2etests.TestServerType,
		Image:   e2etests.TestImage,
		SSHKeys: []string{resSSHKey.TFID() + ".id"},
	}
	resServer.SetRName("lb-server-target")

	resNetwork := &network.RData{
		Name:    "lb-target-test-network",
		IPRange: "10.0.0.0/16",
	}
	resNetwork.SetRName("lb-target-test-network")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_network", resNetwork,
					"testdata/r/hcloud_network_subnet", &network.RDataSubnet{
						NetworkID:   "hcloud_network.lb-target-test-network.id",
						Type:        "cloud",
						NetworkZone: "eu-central",
						IPRange:     "10.0.1.0/24",
					},
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_server", resServer,
					"testdata/r/hcloud_server_network", &server.RDataNetwork{
						Name:      "lb-server-network",
						ServerID:  resServer.TFID() + ".id",
						NetworkID: "hcloud_network.lb-target-test-network.id",
					},
					"testdata/r/hcloud_load_balancer", &loadbalancer.RData{
						Name:        "target-test-lb",
						Type:        e2etests.TestLoadBalancerType,
						NetworkZone: "eu-central",
					},
					"testdata/r/hcloud_load_balancer_network", &loadbalancer.RDataNetwork{
						Name:                  "target-test-lb-network",
						LoadBalancerID:        "hcloud_load_balancer.target-test-lb.id",
						NetworkID:             "hcloud_network.lb-target-test-network.id",
						EnablePublicInterface: true,
					},
					"testdata/r/hcloud_load_balancer_target", &loadbalancer.RDataTarget{
						Name:           "lb-test-target",
						Type:           "server",
						LoadBalancerID: "hcloud_load_balancer.target-test-lb.id",
						ServerID:       resServer.TFID() + ".id",
						UsePrivateIP:   true,
						DependsOn: []string{
							"hcloud_server_network.lb-server-network",
							"hcloud_load_balancer_network.target-test-lb-network",
						},
					},
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(
						loadbalancer.ResourceType+".target-test-lb", loadbalancer.ByID(t, &lb)),
					testsupport.CheckResourceExists(resServer.TFID(), server.ByID(t, &srv)),
					resource.TestCheckResourceAttr(
						loadbalancer.TargetResourceType+".lb-test-target", "use_private_ip", "true"),
					testsupport.LiftTCF(hasServerTarget(&lb, &srv)),
				),
			},
			{
				ResourceName:      loadbalancer.TargetResourceType + ".lb-test-target",
				ImportState:       true,
				ImportStateIdFunc: loadBalancerTargetImportStateIDFunc("target-test-lb", hcloud.LoadBalancerTargetTypeServer, "lb-server-target"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccHcloudLoadBalancerTarget_LabelSelectorTarget(t *testing.T) {
	var (
		lb  hcloud.LoadBalancer
		srv hcloud.Server
	)

	tmplMan := testtemplate.Manager{}
	selector := fmt.Sprintf("tf-test=tf-test-%d", tmplMan.RandInt)
	resSSHKey := sshkey.NewRData(t, "lb-label-target")
	resServer := &server.RData{
		Name:  "lb-server-target",
		Type:  e2etests.TestServerType,
		Image: e2etests.TestImage,
		Labels: map[string]string{
			"tf-test": fmt.Sprintf("tf-test-%d", tmplMan.RandInt),
		},
		SSHKeys: []string{resSSHKey.TFID() + ".id"},
	}
	resServer.SetRName("lb-server-target")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_server", resServer,
					"testdata/r/hcloud_load_balancer", &loadbalancer.RData{
						Name:        "target-test-lb",
						Type:        e2etests.TestLoadBalancerType,
						NetworkZone: "eu-central",
					},
					"testdata/r/hcloud_load_balancer_target", &loadbalancer.RDataTarget{
						Name:           "lb-test-target",
						Type:           "label_selector",
						LoadBalancerID: "hcloud_load_balancer.target-test-lb.id",
						LabelSelector:  selector,
					},
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(
						loadbalancer.ResourceType+".target-test-lb", loadbalancer.ByID(t, &lb)),
					testsupport.CheckResourceExists(resServer.TFID(), server.ByID(t, &srv)),
					resource.TestCheckResourceAttr(
						loadbalancer.TargetResourceType+".lb-test-target", "type", "label_selector"),
					resource.TestCheckResourceAttr(
						loadbalancer.TargetResourceType+".lb-test-target", "label_selector", selector),
					testsupport.LiftTCF(hasLabelSelectorTarget(&lb, selector)),
				),
			},
			{
				ResourceName:      loadbalancer.TargetResourceType + ".lb-test-target",
				ImportState:       true,
				ImportStateIdFunc: loadBalancerTargetImportStateIDFunc("target-test-lb", hcloud.LoadBalancerTargetTypeLabelSelector, selector),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccHcloudLoadBalancerTarget_LabelSelectorTarget_UsePrivateIP(t *testing.T) {
	var (
		lb  hcloud.LoadBalancer
		srv hcloud.Server
	)
	tmplMan := testtemplate.Manager{}

	resSSHKey := sshkey.NewRData(t, "lb-label-target")
	resServer := &server.RData{
		Name:  "lb-server-target",
		Type:  e2etests.TestServerType,
		Image: e2etests.TestImage,
		Labels: map[string]string{
			"tf-test": fmt.Sprintf("tf-test-%d", tmplMan.RandInt),
		},
		SSHKeys: []string{resSSHKey.TFID() + ".id"},
	}
	resServer.SetRName("lb-server-target")

	resNetwork := &network.RData{
		Name:    "lb-target-test-network",
		IPRange: "10.0.0.0/16",
	}
	resNetwork.SetRName("lb-target-test-network")

	resSubNet := &network.RDataSubnet{
		NetworkID:   "hcloud_network.lb-target-test-network.id",
		Type:        "cloud",
		NetworkZone: "eu-central",
		IPRange:     "10.0.1.0/24",
	}
	resSubNet.SetRName("lb-target-test-sub-network")

	selector := fmt.Sprintf("tf-test=tf-test-%d", tmplMan.RandInt)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_network", resNetwork,
					"testdata/r/hcloud_network_subnet", resSubNet,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_server", resServer,
					"testdata/r/hcloud_server_network", &server.RDataNetwork{
						Name:      "lb-server-network",
						ServerID:  resServer.TFID() + ".id",
						NetworkID: "hcloud_network.lb-target-test-network.id",
					},
					"testdata/r/hcloud_load_balancer", &loadbalancer.RData{
						Name:        "target-test-lb",
						Type:        e2etests.TestLoadBalancerType,
						NetworkZone: "eu-central",
					},
					"testdata/r/hcloud_load_balancer_network", &loadbalancer.RDataNetwork{
						Name:                  "target-test-lb-network",
						LoadBalancerID:        "hcloud_load_balancer.target-test-lb.id",
						NetworkID:             "hcloud_network.lb-target-test-network.id",
						EnablePublicInterface: true,
						DependsOn: []string{
							resSubNet.TFID(),
						},
					},
					"testdata/r/hcloud_load_balancer_target", &loadbalancer.RDataTarget{
						Name:           "lb-test-target",
						Type:           "label_selector",
						LoadBalancerID: "hcloud_load_balancer.target-test-lb.id",
						LabelSelector:  selector,
						UsePrivateIP:   true,
						DependsOn: []string{
							"hcloud_server_network.lb-server-network",
							"hcloud_load_balancer_network.target-test-lb-network",
						},
					},
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(
						loadbalancer.ResourceType+".target-test-lb", loadbalancer.ByID(t, &lb)),
					testsupport.CheckResourceExists(
						server.ResourceType+".lb-server-target", server.ByID(t, &srv)),
					resource.TestCheckResourceAttr(
						loadbalancer.TargetResourceType+".lb-test-target", "type", "label_selector"),
					resource.TestCheckResourceAttr(
						loadbalancer.TargetResourceType+".lb-test-target", "label_selector", selector),
					resource.TestCheckResourceAttr(
						loadbalancer.TargetResourceType+".lb-test-target", "use_private_ip", "true"),
					testsupport.LiftTCF(hasLabelSelectorTarget(&lb, selector)),
				),
			},
			{
				ResourceName:      loadbalancer.TargetResourceType + ".lb-test-target",
				ImportState:       true,
				ImportStateIdFunc: loadBalancerTargetImportStateIDFunc("target-test-lb", hcloud.LoadBalancerTargetTypeLabelSelector, selector),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccHcloudLoadBalancerTarget_IPTarget(t *testing.T) {
	var (
		lb hcloud.LoadBalancer
	)

	ip := "213.239.214.25"
	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", &loadbalancer.RData{
						Name:        "target-test-lb",
						Type:        e2etests.TestLoadBalancerType,
						NetworkZone: "eu-central",
					},
					"testdata/r/hcloud_load_balancer_target", &loadbalancer.RDataTarget{
						Name:           "lb-test-target",
						LoadBalancerID: "hcloud_load_balancer.target-test-lb.id",
						Type:           "ip",
						IP:             ip,
					},
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(
						loadbalancer.ResourceType+".target-test-lb", loadbalancer.ByID(t, &lb)),
					resource.TestCheckResourceAttr(
						loadbalancer.TargetResourceType+".lb-test-target", "type", "ip"),
					resource.TestCheckResourceAttr(
						loadbalancer.TargetResourceType+".lb-test-target", "ip", ip),
					testsupport.LiftTCF(hasIPTarget(&lb, ip)),
				),
			},
			{
				ResourceName:      loadbalancer.TargetResourceType + ".lb-test-target",
				ImportState:       true,
				ImportStateIdFunc: loadBalancerTargetImportStateIDFunc("target-test-lb", hcloud.LoadBalancerTargetTypeIP, ip),
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
