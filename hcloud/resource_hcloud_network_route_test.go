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
	resource.AddTestSweepers("hcloud_network_route", &resource.Sweeper{
		Name: "hcloud_network_route",
		F:    testSweepNetworks,
	})
}

func TestAccHcloudNetworkRoute_Basic(t *testing.T) {
	var network hcloud.Network
	var route hcloud.NetworkRoute
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccHcloudPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccHcloudCheckNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckNetworkRouteConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckNetworkExists("hcloud_network.foobar_network_route", &network),
					resource.TestCheckResourceAttr(
						"hcloud_network.foobar_network_route", "name", fmt.Sprintf("foo-network-route-%d", rInt)),
					resource.TestCheckResourceAttr(
						"hcloud_network.foobar_network_route", "ip_range", "10.0.0.0/8"),
					testAccHcloudCheckNetworkRouteExists("hcloud_network_route.fooroute", &route),
					resource.TestCheckResourceAttr(
						"hcloud_network_route.fooroute", "destination", "10.100.1.0/24"),
					resource.TestCheckResourceAttr(
						"hcloud_network_route.fooroute", "gateway", "10.0.1.1"),
				),
			},
		},
	})
}

func testAccHcloudCheckNetworkRouteConfig(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_network" "foobar_network_route" {
  name       = "foo-network-route-%d"
  ip_range   = "10.0.0.0/8"
}
resource "hcloud_network_subnet" "foonet" {
  network_id = "${hcloud_network.foobar_network_route.id}"
  type = "server"
  network_zone = "eu-central"
  ip_range   = "10.0.1.0/24"
}
resource "hcloud_network_route" "fooroute" {
  network_id = "${hcloud_network.foobar_network_route.id}"
  destination   = "10.100.1.0/24"
  gateway   = "10.0.1.1"
}
`, rInt)
}

func testAccHcloudCheckNetworkRouteExists(n string, route *hcloud.NetworkRoute) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		client := testAccProvider.Meta().(*hcloud.Client)
		_, foundRoute, err := lookupNetworkRouteID(context.Background(), rs.Primary.ID, client)
		if err != nil {
			return err
		}
		route = &foundRoute
		return nil
	}
}
