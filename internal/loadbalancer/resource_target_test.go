package loadbalancer_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/loadbalancer"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/network"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/server"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/testsupport"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccHcloudLoadBalancerTarget(t *testing.T) {
	var (
		lb  hcloud.LoadBalancer
		srv hcloud.Server
	)

	tmplMan := testtemplate.Manager{}
	resource.Test(t, resource.TestCase{
		PreCheck:     testsupport.AccTestPreCheck(t),
		Providers:    testsupport.AccTestProviders(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_server", &server.RData{
						Name:  "lb-server-target",
						Type:  "cx11",
						Image: "ubuntu-20.04",
					},
					"testdata/r/hcloud_load_balancer", &loadbalancer.RData{
						Name:        "target-test-lb",
						Type:        "lb11",
						NetworkZone: "eu-central",
					},
					"testdata/r/hcloud_load_balancer_target", &loadbalancer.RDataTarget{
						Name:           "lb-test-target",
						Type:           "server",
						LoadBalancerID: "hcloud_load_balancer.target-test-lb.id",
						ServerID:       "hcloud_server.lb-server-target.id",
					},
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(
						loadbalancer.ResourceType+".target-test-lb", loadbalancer.ByID(t, &lb)),
					testsupport.CheckResourceExists(
						server.ResourceType+".lb-server-target", server.ByID(t, &srv)),
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
		},
	})
}

func TestAccHcloudLoadBalancerTarget_UsePrivateIP(t *testing.T) {
	var (
		lb  hcloud.LoadBalancer
		srv hcloud.Server
	)

	tmplMan := testtemplate.Manager{}
	resource.Test(t, resource.TestCase{
		PreCheck:     testsupport.AccTestPreCheck(t),
		Providers:    testsupport.AccTestProviders(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_network", &network.RData{
						Name:    "lb-target-test-network",
						IPRange: "10.0.0.0/16",
					},
					"testdata/r/hcloud_network_subnet", &network.RDataSubnet{
						Name:        "lb-target-test-subnet",
						NetworkID:   "hcloud_network.lb-target-test-network.id",
						Type:        "cloud",
						NetworkZone: "eu-central",
						IPRange:     "10.0.1.0/24",
					},
					"testdata/r/hcloud_server", &server.RData{
						Name:  "lb-server-target",
						Type:  "cx11",
						Image: "ubuntu-20.04",
					},
					"testdata/r/hcloud_server_network", &server.RDataNetwork{
						Name:      "lb-server-network",
						ServerID:  "hcloud_server.lb-server-target.id",
						NetworkID: "hcloud_network.lb-target-test-network.id",
					},
					"testdata/r/hcloud_load_balancer", &loadbalancer.RData{
						Name:        "target-test-lb",
						Type:        "lb11",
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
						ServerID:       "hcloud_server.lb-server-target.id",
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
						loadbalancer.TargetResourceType+".lb-test-target", "use_private_ip", "true"),
					testsupport.LiftTCF(hasServerTarget(&lb, &srv)),
				),
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
