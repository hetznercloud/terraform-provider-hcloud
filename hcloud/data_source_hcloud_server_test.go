package hcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func init() {
	resource.AddTestSweepers("data_source_server", &resource.Sweeper{
		Name: "hcloud_server_data_source",
		F:    testSweepServers,
	})
}

func TestAccHcloudDataSourceServer(t *testing.T) {
	var server hcloud.Server
	rInt := acctest.RandInt()
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccHcloudPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckServerDataSourceConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckServerExists("hcloud_server.server", &server),
					resource.TestCheckResourceAttr(
						"data.hcloud_server.s_1", "server_type", "cx11"),
					resource.TestCheckResourceAttr(
						"data.hcloud_server.s_1", "name", fmt.Sprintf("Hashi-Test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.hcloud_server.s_1", "backups", "false"),

					resource.TestCheckResourceAttr(
						"data.hcloud_server.s_2", "server_type", "cx11"),
					resource.TestCheckResourceAttr(
						"data.hcloud_server.s_2", "name", fmt.Sprintf("Hashi-Test-%d", rInt)),

					resource.TestCheckResourceAttr(
						"data.hcloud_server.s_3", "server_type", "cx11"),
					resource.TestCheckResourceAttr(
						"data.hcloud_server.s_3", "name", fmt.Sprintf("Hashi-Test-%d", rInt)),
				),
			},
		},
	})
}

func testAccHcloudCheckServerDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
variable "labels" {
  type = "map"
  default = {
    "key" = "%d"
  }
}
resource "hcloud_server" "server" {
  server_type      = "cx11"
  name    = "Hashi-Test-%d"
  labels  = "${var.labels}"
  image   = "ubuntu-18.04"
}
data "hcloud_server" "s_1" {
  name = "${hcloud_server.server.name}"
}
data "hcloud_server" "s_2" {
  id = "${hcloud_server.server.id}"
}
data "hcloud_server" "s_3" {
  with_selector =  "key=${hcloud_server.server.labels["key"]}"
  with_status = ["running","starting"]
}
`, rInt, rInt)
}
