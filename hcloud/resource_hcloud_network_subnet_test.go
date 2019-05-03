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
	resource.AddTestSweepers("hcloud_network_subnet", &resource.Sweeper{
		Name: "hcloud_network_subnet",
		F:    testSweepNetworks,
	})
}

func TestAccHcloudNetworkSubnet_Basic(t *testing.T) {
	var network hcloud.Network
	var subnet hcloud.NetworkSubnet
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccHcloudPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccHcloudCheckNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckNetworkSubnetConfigServer(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckNetworkExists("hcloud_network.foobar_network", &network),
					resource.TestCheckResourceAttr(
						"hcloud_network.foobar_network", "name", fmt.Sprintf("foo-network-%d", rInt)),
					resource.TestCheckResourceAttr(
						"hcloud_network.foobar_network", "ip_range", "10.0.0.0/16"),
					testAccHcloudCheckNetworkSubnetExists("hcloud_network_subnet.foonet", &subnet),
					resource.TestCheckResourceAttr(
						"hcloud_network_subnet.foonet", "type", "server"),
					resource.TestCheckResourceAttr(
						"hcloud_network_subnet.foonet", "network_zone", "eu-central"),
					resource.TestCheckResourceAttr(
						"hcloud_network_subnet.foonet", "ip_range", "10.0.1.0/24"),
					resource.TestCheckResourceAttr(
						"hcloud_network_subnet.foonet", "gateway", "10.0.0.1"),
				),
			},
			{
				Config: testAccHcloudCheckNetworkSubnetConfigVSwitch(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckNetworkExists("hcloud_network.foobar_network", &network),
					resource.TestCheckResourceAttr(
						"hcloud_network.foobar_network", "name", fmt.Sprintf("foo-network-%d", rInt)),
					resource.TestCheckResourceAttr(
						"hcloud_network.foobar_network", "ip_range", "10.0.0.0/16"),
					testAccHcloudCheckNetworkSubnetExists("hcloud_network_subnet.foonet_vswitch", &subnet),
					resource.TestCheckResourceAttr(
						"hcloud_network_subnet.foonet_vswitch", "type", "vswitch"),
					resource.TestCheckResourceAttr(
						"hcloud_network_subnet.foonet_vswitch", "network_zone", "eu-central"),
					resource.TestCheckResourceAttr(
						"hcloud_network_subnet.foonet_vswitch", "ip_range", "10.0.100.0/24"),
					resource.TestCheckResourceAttr(
						"hcloud_network_subnet.foonet_vswitch", "gateway", "10.0.100.1"),
				),
			},
		},
	})
}

func testAccHcloudCheckNetworkSubnetConfigServer(rInt int) string {
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
`, rInt)
}
func testAccHcloudCheckNetworkSubnetConfigVSwitch(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_network" "foobar_network" {
  name       = "foo-network-%d"
  ip_range   = "10.0.0.0/16"
}
resource "hcloud_network_subnet" "foonet_vswitch" {
  network_id = "${hcloud_network.foobar_network.id}"
  type = "vswitch"
  network_zone = "eu-central"
  ip_range   = "10.0.100.0/24"
  vswitch_id = 333
}
`, rInt)
}

func testAccHcloudCheckNetworkSubnetExists(n string, subnet *hcloud.NetworkSubnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		client := testAccProvider.Meta().(*hcloud.Client)
		_, foundSubnet, err := lookupNetworkSubnetID(context.Background(), rs.Primary.ID, client)
		if err != nil {
			return err
		}

		*subnet = *foundSubnet
		return nil
	}
}
