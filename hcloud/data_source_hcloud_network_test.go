package hcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func TestAccHcloudDataSourceNetwork(t *testing.T) {
	var network hcloud.Network
	rInt := acctest.RandInt()
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccHcloudPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckNetworkDataSourceConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckNetworkExists("hcloud_network.network_ds", &network),
					resource.TestCheckResourceAttr(
						"data.hcloud_network.network_1", "name", fmt.Sprintf("network-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.hcloud_network.network_1", "ip_range", "10.0.0.0/16"),
					resource.TestCheckResourceAttr(
						"data.hcloud_network.network_2", "name", fmt.Sprintf("network-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.hcloud_network.network_2", "ip_range", "10.0.0.0/16"),
					resource.TestCheckResourceAttr(
						"data.hcloud_network.network_3", "name", fmt.Sprintf("network-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.hcloud_network.network_3", "ip_range", "10.0.0.0/16"),
				),
			},
		},
	})
}
func testAccHcloudCheckNetworkDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
variable "labels" {
  type = "map"
  default = {
    "key" = "%d"
  }
}
resource "hcloud_network" "network_ds" {
  name       = "network-%d"
  ip_range   = "10.0.0.0/16"
  labels     = "${var.labels}"
}
data "hcloud_network" "network_1" {
  name = "${hcloud_network.network_ds.name}"
}
data "hcloud_network" "network_2" {
  id =  "${hcloud_network.network_ds.id}"
}
data "hcloud_network" "network_3" {
  with_selector =  "key=${hcloud_network.network_ds.labels["key"]}"
}
`, rInt, rInt)
}

