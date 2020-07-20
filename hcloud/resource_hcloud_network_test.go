package hcloud

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func init() {
	resource.AddTestSweepers("hcloud_network", &resource.Sweeper{
		Name: "hcloud_network",
		F:    testSweepNetworks,
	})
}

func TestAccHcloudNetwork_Basic(t *testing.T) {
	var network hcloud.Network
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccHcloudPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccHcloudCheckNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckNetworkConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckNetworkExists("hcloud_network.foobar", &network),
					resource.TestCheckResourceAttr(
						"hcloud_network.foobar", "name", fmt.Sprintf("foo-network-%d", rInt)),
					resource.TestCheckResourceAttr(
						"hcloud_network.foobar", "ip_range", "10.0.0.0/16"),
				),
			},
			{
				Config: testAccHcloudCheckNetworkConfig_Renamed(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckNetworkExists("hcloud_network.foobar", &network),
					resource.TestCheckResourceAttr(
						"hcloud_network.foobar", "name", fmt.Sprintf("foo-network-renamed-%d", rInt)),
					resource.TestCheckResourceAttr(
						"hcloud_network.foobar", "ip_range", "10.0.0.0/16"),
				),
			},
		},
	})
}

func testAccHcloudCheckNetworkConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_network" "foobar" {
  name       = "foo-network-%d"
  ip_range   = "10.0.0.0/16"
}
`, rInt)
}
func testAccHcloudCheckNetworkConfig_Renamed(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_network" "foobar" {
  name       = "foo-network-renamed-%d"
  ip_range   = "10.0.0.0/16"
}
`, rInt)
}

func testAccHcloudCheckNetworkDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*hcloud.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "hcloud_network" {
			continue
		}

		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("network id is no int: %v", err)
		}
		var network *hcloud.Network
		network, _, err = client.Network.GetByID(context.Background(), id)
		if err != nil {
			return fmt.Errorf(
				"Error checking if network (%s) is deleted: %v",
				rs.Primary.ID, err)
		}
		if network != nil {
			return fmt.Errorf("network (%s) has not been deleted", rs.Primary.ID)
		}
	}

	return nil
}
func testAccHcloudCheckNetworkExists(n string, network *hcloud.Network) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		client := testAccProvider.Meta().(*hcloud.Client)
		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return err
		}

		// Try to find the key
		foundNetwork, _, err := client.Network.GetByID(context.Background(), id)
		if err != nil {
			return err
		}

		if foundNetwork == nil {
			return fmt.Errorf("Record not found")
		}

		*network = *foundNetwork
		return nil
	}
}

// Deprecated: use network.Sweep instead
func testSweepNetworks(region string) error {
	client, err := createClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	networks, err := client.Network.All(ctx)
	if err != nil {
		return err
	}

	for _, network := range networks {
		if _, err := client.Network.Delete(ctx, network); err != nil {
			return err
		}
	}

	return nil
}
