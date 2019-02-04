package hcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func init() {
	resource.AddTestSweepers("data_source_datacenter", &resource.Sweeper{
		Name: "hcloud_datacenter_data_source",
	})
}
func TestAccHcloudDataSourceDatasource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccHcloudPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckDatacenterDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.hcloud_datacenter.ds_1", "id", "4"),
					resource.TestCheckResourceAttr(
						"data.hcloud_datacenter.ds_1", "name", "fsn1-dc14"),
					resource.TestCheckResourceAttr(
						"data.hcloud_datacenter.ds_1", "description", "Falkenstein 1 DC14"),
					resource.TestCheckResourceAttr(
						"data.hcloud_datacenter.ds_2", "id", "4"),
					resource.TestCheckResourceAttr(
						"data.hcloud_datacenter.ds_2", "name", "fsn1-dc14"),
					resource.TestCheckResourceAttr(
						"data.hcloud_datacenter.ds_2", "description", "Falkenstein 1 DC14"),
				),
			},
		},
	})
}

func testAccHcloudCheckDatacenterDataSourceConfig() string {
	return fmt.Sprintf(`
data "hcloud_datacenter" "ds_1" {
  name = "fsn1-dc14"
}
data "hcloud_datacenter" "ds_2" {
  id = 4
}
`)
}
