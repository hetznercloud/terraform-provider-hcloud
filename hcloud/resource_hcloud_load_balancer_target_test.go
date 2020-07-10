package hcloud

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func TestAccHcloudLoadBalancerTarget(t *testing.T) {
	var (
		lb  hcloud.LoadBalancer
		srv hcloud.Server
	)

	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccHcloudPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudLoadBalancerTarget(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckLoadBalancerExists("hcloud_load_balancer.target_test_lb", &lb),
					testAccHcloudCheckServerExists("hcloud_server.lb_server_target", &srv),
					testAccHcloudLoadBalancerTargetHasServerTarget("lb_test_target", &srv, &lb),
				),
			},
			{

				Config: testAccHcloudLoadBalancerTarget_UsePrivateIP(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckLoadBalancerExists("hcloud_load_balancer.target_test_lb", &lb),
					testAccHcloudCheckServerExists("hcloud_server.lb_server_target", &srv),
					testAccHcloudLoadBalancerTargetHasServerTarget("lb_test_target", &srv, &lb),
					resource.TestCheckResourceAttr("hcloud_load_balancer_target.lb_test_target", "use_private_ip", "true"),
				),
			},
		},
	})
}

func testAccHcloudLoadBalancerTarget(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_server" "lb_server_target" {
	name        = "lb-server-target-%d"
	server_type = "cx11"
	image       = "ubuntu-18.04"
}

resource "hcloud_load_balancer" "target_test_lb" {
	name               = "target-test-lb-%d"
	load_balancer_type = "lb11"
	network_zone       = "eu-central"

	algorithm {
		type = "round_robin"
	}
}

resource "hcloud_load_balancer_target" "lb_test_target" {
	type             = "server"
	load_balancer_id = "${hcloud_load_balancer.target_test_lb.id}"
	server_id        = "${hcloud_server.lb_server_target.id}"
}
	`, rInt, rInt)
}

func testAccHcloudLoadBalancerTarget_UsePrivateIP(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_network" "lb_target_test_network" {
	name         = "lb-target-test-network-%d"
	ip_range     = "10.0.0.0/16"
}

resource "hcloud_network_subnet" "lb_target_test_subnet" {
	network_id   = "${hcloud_network.lb_target_test_network.id}"
	type         = "cloud"
	network_zone = "eu-central"
	ip_range     = "10.0.1.0/24"
}

resource "hcloud_server" "lb_server_target" {
	name        = "lb-server-target-%d"
	server_type = "cx11"
	image       = "ubuntu-18.04"
}

resource "hcloud_server_network" "lb_server_network" {
	server_id  = "${hcloud_server.lb_server_target.id}"
	network_id = "${hcloud_network.lb_target_test_network.id}"
	depends_on = [hcloud_network_subnet.lb_target_test_subnet]
}

resource "hcloud_load_balancer" "target_test_lb" {
	name               = "target-test-lb-%d"
	load_balancer_type = "lb11"
	network_zone       = "eu-central"

	algorithm {
		type = "round_robin"
	}
}

resource "hcloud_load_balancer_network" "target_test_lb_network" {
	load_balancer_id        = "${hcloud_load_balancer.target_test_lb.id}"
	network_id              = "${hcloud_network.lb_target_test_network.id}"
	enable_public_interface = true
	depends_on              = [hcloud_network_subnet.lb_target_test_subnet]
}

resource "hcloud_load_balancer_target" "lb_test_target" {
	type             = "server"
	load_balancer_id = "${hcloud_load_balancer.target_test_lb.id}"
	server_id        = "${hcloud_server.lb_server_target.id}"
	use_private_ip   = true

	depends_on = [
		hcloud_load_balancer_network.target_test_lb_network,
		hcloud_server_network.lb_server_network
	]
}
	`, rInt, rInt, rInt)
}

func testAccHcloudLoadBalancerTargetHasServerTarget(
	name string, srv *hcloud.Server, lb *hcloud.LoadBalancer,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resName := fmt.Sprintf("hcloud_load_balancer_target.%s", name)
		if err := resource.TestCheckResourceAttr(resName, "type", "server")(s); err != nil {
			return err
		}
		if err := resource.TestCheckResourceAttr(resName, "load_balancer_id", strconv.Itoa(lb.ID))(s); err != nil {
			return err
		}
		if err := resource.TestCheckResourceAttr(resName, "server_id", strconv.Itoa(srv.ID))(s); err != nil {
			return err
		}

		for _, tgt := range lb.Targets {
			if tgt.Type == hcloud.LoadBalancerTargetTypeServer && tgt.Server.Server.ID == srv.ID {
				return nil
			}
		}

		return fmt.Errorf("load balancer has no target for server with id: %d", srv.ID)
	}
}
