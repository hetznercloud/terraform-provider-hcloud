package hcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func TestAccHcloudDataSourceVolume(t *testing.T) {
	var volume hcloud.Volume
	rInt := acctest.RandInt()
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccHcloudPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckVolumeDataSourceConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckVolumeExists("hcloud_volume.volume_ds", &volume),
					resource.TestCheckResourceAttr(
						"data.hcloud_volume.volume_1", "name", fmt.Sprintf("volume-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.hcloud_volume.volume_2", "name", fmt.Sprintf("volume-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.hcloud_volume.volume_3", "name", fmt.Sprintf("volume-%d", rInt)),

				),
			},
		},
	})
}
func testAccHcloudCheckVolumeDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
variable "labels" {
  type = "map"
  default = {
    "key" = "value"
  }
}
resource "hcloud_volume" "volume_ds" {
  name    = "volume-%d"
  size    = 123
  labels  = "${var.labels}"
}
data "hcloud_volume" "volume_1" {
  name = "${hcloud_volume.volume_ds.name}"
}
data "hcloud_volume" "volume_2" {
  id =  "${hcloud_volume.volume_ds.id}"
}
data "hcloud_volume" "volume_3" {
  selector =  "key=${hcloud_volume.volume_ds.labels["key"]}"
}
`, rInt)
}
