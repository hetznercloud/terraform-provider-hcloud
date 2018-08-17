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

var floatingIPForDataSource *hcloud.FloatingIP

func TestAccHcloudDataSourceFloatingIP(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccHcloudPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckFloatingIPDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckFloatingIPExists("data.hcloud_floating_ip.ip_1", floatingIPForDataSource),
					resource.TestCheckResourceAttr(
						"data.hcloud_floating_ip.ip_1", "type", "ipv4"),
					resource.TestCheckResourceAttr(
						"data.hcloud_floating_ip.ip_1", "home_location", "fsn1"),
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
}
data "hcloud_floating_ip" "ip_1" {
  ip_address = "${hcloud_floating_ip.floating_ip.ip_address}"
}`)
}

func testDataSourceCleanup() {
	testSweepFloatingIps("all")
}
