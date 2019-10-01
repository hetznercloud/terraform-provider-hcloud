package hcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func init() {
	resource.AddTestSweepers("data_source_locations", &resource.Sweeper{
		Name: "hcloud_locations_data_source",
	})
}
func TestAccHcloudDataSourceLocations(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccHcloudPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckLocationsDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.hcloud_locations.l", "location_ids.0", "1"),
					resource.TestCheckResourceAttr(
						"data.hcloud_locations.l", "location_ids.1", "2"),
					resource.TestCheckResourceAttr(
						"data.hcloud_locations.l", "location_ids.2", "3"),
					resource.TestCheckResourceAttr(
						"data.hcloud_locations.l", "names.0", "fsn1"),
					resource.TestCheckResourceAttr(
						"data.hcloud_locations.l", "names.1", "nbg1"),
					resource.TestCheckResourceAttr(
						"data.hcloud_locations.l", "names.2", "hel1"),
					resource.TestCheckResourceAttr(
						"data.hcloud_locations.l", "descriptions.0", "Falkenstein DC Park 1"),
					resource.TestCheckResourceAttr(
						"data.hcloud_locations.l", "descriptions.1", "Nuremberg DC Park 1"),
					resource.TestCheckResourceAttr(
						"data.hcloud_locations.l", "descriptions.2", "Helsinki DC Park 1"),
				),
			},
		},
	})
}

func testAccHcloudCheckLocationsDataSourceConfig() string {
	return fmt.Sprintf(`
data "hcloud_locations" "l" {
}
`)
}
