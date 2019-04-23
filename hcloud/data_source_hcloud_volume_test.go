package hcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
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
						"data.hcloud_volume.volume_1", "size", "10"),
					resource.TestCheckResourceAttr(
						"data.hcloud_volume.volume_1", "location", "nbg1"),
					testAccHcloudCheckVolumeDataSourceLinuxDevice("data.hcloud_volume.volume_1", &volume),
					resource.TestCheckResourceAttr(
						"data.hcloud_volume.volume_2", "name", fmt.Sprintf("volume-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.hcloud_volume.volume_2", "size", "10"),
					resource.TestCheckResourceAttr(
						"data.hcloud_volume.volume_2", "location", "nbg1"),
					testAccHcloudCheckVolumeDataSourceLinuxDevice("data.hcloud_volume.volume_2", &volume),
					resource.TestCheckResourceAttr(
						"data.hcloud_volume.volume_3", "name", fmt.Sprintf("volume-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.hcloud_volume.volume_3", "size", "10"),
					resource.TestCheckResourceAttr(
						"data.hcloud_volume.volume_3", "location", "nbg1"),
					testAccHcloudCheckVolumeDataSourceLinuxDevice("data.hcloud_volume.volume_3", &volume),
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
    "key" = "%d"
  }
}
resource "hcloud_volume" "volume_ds" {
  name       = "volume-%d"
  size       = 10
  labels     = "${var.labels}"
  location   = "nbg1"
}
data "hcloud_volume" "volume_1" {
  name = "${hcloud_volume.volume_ds.name}"
}
data "hcloud_volume" "volume_2" {
  id =  "${hcloud_volume.volume_ds.id}"
}
data "hcloud_volume" "volume_3" {
  with_selector =  "key=${hcloud_volume.volume_ds.labels["key"]}"
}
`, rInt, rInt)
}

func testAccHcloudCheckVolumeDataSourceLinuxDevice(n string, volume *hcloud.Volume) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}
		linuxDevice := rs.Primary.Attributes["linux_device"]

		if linuxDevice != fmt.Sprintf("/dev/disk/by-id/scsi-0HC_Volume_%d", volume.ID) {
			return fmt.Errorf("Invalid Linux Device on volume: Got %s instead of %s", linuxDevice, fmt.Sprintf("/dev/disk/by-id/scsi-0HC_Volume_%d", volume.ID))
		}

		return nil
	}
}
