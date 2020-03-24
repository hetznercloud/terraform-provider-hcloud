package hcloud

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func init() {
	resource.AddTestSweepers("hcloud_load_balancer_network", &resource.Sweeper{
		Name: "hcloud_load_balancer_network",
		F:    testSweepNetworks,
	})
}

func TestAccHcloudLoadBalancerNetwork(t *testing.T) {
	var network hcloud.Network
	var subnet hcloud.NetworkSubnet
	var lbNet hcloud.LoadBalancerPrivateNet
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccHcloudPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccHcloudCheckNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckLoadBalancerNetwork(rInt, false),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckNetworkExists("hcloud_network.foobar_network", &network),
					resource.TestCheckResourceAttr(
						"hcloud_network.foobar_network", "name", fmt.Sprintf("foo-network-%d", rInt)),
					resource.TestCheckResourceAttr("hcloud_network.foobar_network", "ip_range", "10.0.0.0/16"),
					testAccHcloudCheckNetworkSubnetExists("hcloud_network_subnet.foonet", subnet),
					testAccHcloudCheckLoadBalancerNetworkExists("hcloud_load_balancer_network.lbnetwork", &lbNet),
					resource.TestCheckResourceAttr("hcloud_load_balancer_network.lbnetwork", "ip", "10.0.1.5"),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer_network.lbnetwork", "enable_public_interface", "false"),
				),
			},
			{
				Config: testAccHcloudCheckLoadBalancerNetwork(rInt, true),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckNetworkExists("hcloud_network.foobar_network", &network),
					resource.TestCheckResourceAttr(
						"hcloud_network.foobar_network", "name", fmt.Sprintf("foo-network-%d", rInt)),
					resource.TestCheckResourceAttr("hcloud_network.foobar_network", "ip_range", "10.0.0.0/16"),
					testAccHcloudCheckNetworkSubnetExists("hcloud_network_subnet.foonet", subnet),
					testAccHcloudCheckLoadBalancerNetworkExists("hcloud_load_balancer_network.lbnetwork", &lbNet),
					resource.TestCheckResourceAttr("hcloud_load_balancer_network.lbnetwork", "ip", "10.0.1.5"),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer_network.lbnetwork", "enable_public_interface", "true"),
				),
			},
		},
	})
}

func testAccHcloudCheckLoadBalancerNetwork(rInt int, enablePubIface bool) string {
	return fmt.Sprintf(`
resource "hcloud_network" "foobar_network" {
  name       = "foo-network-%d"
  ip_range   = "10.0.0.0/16"
}

resource "hcloud_network_subnet" "foonet" {
  network_id   = "${hcloud_network.foobar_network.id}"
  type         = "cloud"
  network_zone = "eu-central"
  ip_range     = "10.0.1.0/24"
}

resource "hcloud_load_balancer" "foo" {
  load_balancer_type = "lb11"
  name               = "lb-network-test-%d"
  network_zone       = "eu-central"
}

resource "hcloud_load_balancer_network" "lbnetwork" {
  load_balancer_id        = "${hcloud_load_balancer.foo.id}"
  network_id              = "${hcloud_network.foobar_network.id}"
  ip                      = "10.0.1.5"
  enable_public_interface = %t
}
`, rInt, rInt, enablePubIface)
}

func testAccHcloudCheckLoadBalancerNetworkExists(n string, lbNet *hcloud.LoadBalancerPrivateNet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		client := testAccProvider.Meta().(*hcloud.Client)
		_, _, foundSrvNet, err := lookupLoadBalancerNetworkID(context.Background(), rs.Primary.ID, client)
		if err != nil {
			return err
		}

		*lbNet = *foundSrvNet
		return nil
	}
}
