package hcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func init() {
	resource.AddTestSweepers("data_source_floating_ip", &resource.Sweeper{
		Name: "hcloud_floating_ip_data_source",
		F:    testSweepFloatingIps,
	})
}

func TestAccHcloudDataSourceFloatingIP(t *testing.T) {
	var floatingIP hcloud.FloatingIP
	rInt := acctest.RandInt()
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccHcloudPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckFloatingIPDataSourceConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckFloatingIPExists("data.hcloud_floating_ip.ip_1", &floatingIP),
					resource.TestCheckResourceAttr(
						"data.hcloud_floating_ip.ip_1", "type", "ipv4"),
					resource.TestCheckResourceAttr(
						"data.hcloud_floating_ip.ip_1", "home_location", "fsn1"),
					resource.TestCheckResourceAttr(
						"data.hcloud_floating_ip.ip_1", "description", fmt.Sprintf("Hashi-Test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.hcloud_floating_ip.ip_1", "name", fmt.Sprintf("Hashi-Test-%d", rInt)),

					resource.TestCheckResourceAttr(
						"data.hcloud_floating_ip.ip_2", "type", "ipv4"),
					resource.TestCheckResourceAttr(
						"data.hcloud_floating_ip.ip_2", "home_location", "fsn1"),
					resource.TestCheckResourceAttr(
						"data.hcloud_floating_ip.ip_2", "description", fmt.Sprintf("Hashi-Test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.hcloud_floating_ip.ip_2", "name", fmt.Sprintf("Hashi-Test-%d", rInt)),

					resource.TestCheckResourceAttr(
						"data.hcloud_floating_ip.ip_3", "type", "ipv4"),
					resource.TestCheckResourceAttr(
						"data.hcloud_floating_ip.ip_3", "home_location", "fsn1"),
					resource.TestCheckResourceAttr(
						"data.hcloud_floating_ip.ip_3", "description", fmt.Sprintf("Hashi-Test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.hcloud_floating_ip.ip_3", "name", fmt.Sprintf("Hashi-Test-%d", rInt)),
				),
			},
		},
	})

	testDataSourceCleanup()
}

func testAccHcloudCheckFloatingIPDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
variable "labels" {
  type = "map"
  default = {
    "key" = "%d"
  }
}
resource "hcloud_floating_ip" "floating_ip" {
  type      = "ipv4"
  name      = "Hashi-Test-%d"
  home_location = "fsn1"
  description = "Hashi-Test-%d"
  labels  = "${var.labels}"
}
data "hcloud_floating_ip" "ip_1" {
  ip_address = "${hcloud_floating_ip.floating_ip.ip_address}"
}
data "hcloud_floating_ip" "ip_2" {
  id = "${hcloud_floating_ip.floating_ip.id}"
}
data "hcloud_floating_ip" "ip_3" {
  with_selector =  "key=${hcloud_floating_ip.floating_ip.labels["key"]}"
}
`, rInt, rInt, rInt)
}

func testDataSourceCleanup() {
	testSweepFloatingIps("all")
}
