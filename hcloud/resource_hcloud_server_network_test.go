package hcloud

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/terraform"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func init() {
	resource.AddTestSweepers("hcloud_server_network", &resource.Sweeper{
		Name: "hcloud_server_network",
		F:    testSweepNetworks,
	})
}

func TestAccHcloudServerNetwork(t *testing.T) {
	var network hcloud.Network
	var subnet hcloud.NetworkSubnet
	var srvNet hcloud.ServerPrivateNet
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccHcloudPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccHcloudCheckNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckServerNetwork(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckNetworkExists("hcloud_network.foobar_network", &network),
					resource.TestCheckResourceAttr(
						"hcloud_network.foobar_network", "name", fmt.Sprintf("foo-network-%d", rInt)),
					resource.TestCheckResourceAttr(
						"hcloud_network.foobar_network", "ip_range", "10.0.0.0/16"),
					testAccHcloudCheckNetworkSubnetExists("hcloud_network_subnet.foonet", &subnet),
					testAccHcloudCheckServerNetworkExists("hcloud_server_network.srvnetwork", &srvNet),
					resource.TestCheckResourceAttr(
						"hcloud_server_network.srvnetwork", "io", "10.0.1.5"),
				),
			},
		},
	})
}

func testAccHcloudCheckServerNetwork(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_network" "foobar_network" {
  name       = "foo-network-%d"
  ip_range   = "10.0.0.0/16"
}
resource "hcloud_network_subnet" "foonet" {
  network_id = "${hcloud_network.foobar_network.id}"
  type = "server"
  network_zone = "eu-central"
  ip_range   = "10.0.1.0/24"
}
resource "hcloud_server" "foo" {
  server_type  = "cx11"
  name    = "srv-network-test-%d"
  image   = "ubuntu-18.04"
}
resource "hcloud_server_network" "srvnetwork" {
  server_id = "${hcloud_server.foo.id}"
  network_id = "${hcloud_network.foobar_network.id}"
  ip = "10.0.1.5"
}
`, rInt, rInt)
}

func testAccHcloudCheckServerNetworkExists(n string, srvNet *hcloud.ServerPrivateNet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		client := testAccProvider.Meta().(*hcloud.Client)
		_, _, foundSrvNet, err := lookupServerNetworkID(context.Background(), rs.Primary.ID, client)
		if err != nil {
			return err
		}

		*srvNet = *foundSrvNet
		return nil
	}
}
