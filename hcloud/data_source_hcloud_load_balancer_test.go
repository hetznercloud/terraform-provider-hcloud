package hcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func init() {
	resource.AddTestSweepers("data_source_load_balancer", &resource.Sweeper{
		Name: "hcloud_load_balancer_data_source",
		F:    testSweepLoadBalancers,
	})
}
func TestAccHcloudDataSourceLoadBalancerTest(t *testing.T) {
	var loadBalancer hcloud.LoadBalancer
	rInt := acctest.RandInt()
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccHcloudPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckLoadBalancerDataSourceConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckLoadBalancerExists("hcloud_load_balancer.loadBalancer_ds", &loadBalancer),
					resource.TestCheckResourceAttr(
						"data.hcloud_load_balancer.loadBalancer_1", "name", fmt.Sprintf("loadBalancer-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.hcloud_load_balancer.loadBalancer_1", "location", "nbg1"),
					resource.TestCheckResourceAttr(
						"data.hcloud_load_balancer.loadBalancer_1", "target.#", "0"),
					resource.TestCheckResourceAttr(
						"data.hcloud_load_balancer.loadBalancer_1", "service.#", "0"),
					resource.TestCheckResourceAttr(
						"data.hcloud_load_balancer.loadBalancer_2", "name", fmt.Sprintf("loadBalancer-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.hcloud_load_balancer.loadBalancer_2", "location", "nbg1"),
					resource.TestCheckResourceAttr(
						"data.hcloud_load_balancer.loadBalancer_3", "name", fmt.Sprintf("loadBalancer-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.hcloud_load_balancer.loadBalancer_3", "location", "nbg1"),
				),
			},
		},
	})
}
func testAccHcloudCheckLoadBalancerDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
variable "labels" {
  type = "map"
  default = {
    "key" = "%d"
  }
}
resource "hcloud_load_balancer" "loadBalancer_ds" {
  name       = "loadBalancer-%d"
  location   = "nbg1"
  load_balancer_type = "lb11"
  labels     = "${var.labels}"
}
data "hcloud_load_balancer" "loadBalancer_1" {
  name = "${hcloud_load_balancer.loadBalancer_ds.name}"
}
data "hcloud_load_balancer" "loadBalancer_2" {
  id =  "${hcloud_load_balancer.loadBalancer_ds.id}"
}
data "hcloud_load_balancer" "loadBalancer_3" {
  with_selector =  "key=${hcloud_load_balancer.loadBalancer_ds.labels["key"]}"
}
`, rInt, rInt)
}
