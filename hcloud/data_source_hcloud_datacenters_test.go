package hcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func init() {
	resource.AddTestSweepers("data_source_datacenters", &resource.Sweeper{
		Name: "hcloud_datacenters_data_source",
	})
}
func TestAccHcloudDataSourceDatasources(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccHcloudPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckDatacentersDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.hcloud_datacenters.ds", "datacenter_ids.0", "1"),
					resource.TestCheckResourceAttr(
						"data.hcloud_datacenters.ds", "datacenter_ids.1", "2"),
					resource.TestCheckResourceAttr(
						"data.hcloud_datacenters.ds", "datacenter_ids.2", "3"),
					resource.TestCheckResourceAttr(
						"data.hcloud_datacenters.ds", "datacenter_ids.3", "4"),
					resource.TestCheckResourceAttr(
						"data.hcloud_datacenters.ds", "names.0", "fsn1-dc8"),
					resource.TestCheckResourceAttr(
						"data.hcloud_datacenters.ds", "names.1", "nbg1-dc3"),
					resource.TestCheckResourceAttr(
						"data.hcloud_datacenters.ds", "names.2", "hel1-dc2"),
					resource.TestCheckResourceAttr(
						"data.hcloud_datacenters.ds", "names.3", "fsn1-dc14"),
					resource.TestCheckResourceAttr(
						"data.hcloud_datacenters.ds", "descriptions.0", "Falkenstein 1 DC 8"),
					resource.TestCheckResourceAttr(
						"data.hcloud_datacenters.ds", "descriptions.1", "Nuremberg 1 DC 3"),
					resource.TestCheckResourceAttr(
						"data.hcloud_datacenters.ds", "descriptions.2", "Helsinki 1 DC 2"),
					resource.TestCheckResourceAttr(
						"data.hcloud_datacenters.ds", "descriptions.3", "Falkenstein 1 DC14"),
				),
			},
		},
	})
}

func testAccHcloudCheckDatacentersDataSourceConfig() string {
	return fmt.Sprintf(`
data "hcloud_datacenters" "ds" {
}
`)
}
