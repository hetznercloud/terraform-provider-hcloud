package hcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func init() {
	resource.AddTestSweepers("data_source_location", &resource.Sweeper{
		Name: "hcloud_location_data_source",
	})
}
func TestAccHcloudDataSourceLocation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccHcloudPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckLocationDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.hcloud_location.l_1", "id", "1"),
					resource.TestCheckResourceAttr(
						"data.hcloud_location.l_1", "name", "fsn1"),
					resource.TestCheckResourceAttr(
						"data.hcloud_location.l_1", "description", "Falkenstein DC Park 1"),
					resource.TestCheckResourceAttr(
						"data.hcloud_location.l_2", "id", "1"),
					resource.TestCheckResourceAttr(
						"data.hcloud_location.l_2", "name", "fsn1"),
					resource.TestCheckResourceAttr(
						"data.hcloud_location.l_2", "description", "Falkenstein DC Park 1"),
				),
			},
		},
	})
}

func testAccHcloudCheckLocationDataSourceConfig() string {
	return fmt.Sprintf(`
data "hcloud_location" "l_1" {
  name = "fsn1"
}
data "hcloud_location" "l_2" {
  id = 1
}
`)
}
