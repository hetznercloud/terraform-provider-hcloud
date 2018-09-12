package hcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func init() {
	resource.AddTestSweepers("data_source_image", &resource.Sweeper{
		Name: "hcloud_image_data_source",
	})
}
func TestAccHcloudDataSourceImage(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccHcloudPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckImageDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.hcloud_image.image_1", "type", "system"),
					resource.TestCheckResourceAttr(
						"data.hcloud_image.image_1", "os_flavor", "ubuntu"),
					resource.TestCheckResourceAttr(
						"data.hcloud_image.image_1", "description", "Ubuntu 18.04"),
					resource.TestCheckResourceAttr(
						"data.hcloud_image.image_2", "type", "system"),
					resource.TestCheckResourceAttr(
						"data.hcloud_image.image_2", "os_flavor", "ubuntu"),
					resource.TestCheckResourceAttr(
						"data.hcloud_image.image_2", "description", "Ubuntu 18.04"),
				),
			},
		},
	})
}
func testAccHcloudCheckImageDataSourceConfig() string {
	return fmt.Sprintf(`
data "hcloud_image" "image_1" {
  name = "ubuntu-18.04"
}
data "hcloud_image" "image_2" {
  id = 168855
}
`)
}
