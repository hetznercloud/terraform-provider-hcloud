package hcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
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
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccHcloudPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckFloatingIPDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckFloatingIPExists("data.hcloud_floating_ip.ip_1", &floatingIP),
					resource.TestCheckResourceAttr(
						"data.hcloud_floating_ip.ip_1", "type", "ipv4"),
					resource.TestCheckResourceAttr(
						"data.hcloud_floating_ip.ip_1", "home_location", "fsn1"),
					resource.TestCheckResourceAttr(
						"data.hcloud_floating_ip.ip_1", "description", "Hashi Test"),

					resource.TestCheckResourceAttr(
						"data.hcloud_floating_ip.ip_2", "type", "ipv4"),
					resource.TestCheckResourceAttr(
						"data.hcloud_floating_ip.ip_2", "home_location", "fsn1"),
					resource.TestCheckResourceAttr(
						"data.hcloud_floating_ip.ip_2", "description", "Hashi Test"),
				),
			},
		},
	})

	testDataSourceCleanup()
}

func testAccHcloudCheckFloatingIPDataSourceConfig() string {
	return fmt.Sprintf(`
resource "hcloud_floating_ip" "floating_ip" {
  type      = "ipv4"
  home_location = "fsn1"
	description = "Hashi Test"
}
data "hcloud_floating_ip" "ip_1" {
  ip_address = "${hcloud_floating_ip.floating_ip.ip_address}"
}
data "hcloud_floating_ip" "ip_2" {
  id = "${hcloud_floating_ip.floating_ip.id}"
}`)
}

func testDataSourceCleanup() {
	testSweepFloatingIps("all")
}
