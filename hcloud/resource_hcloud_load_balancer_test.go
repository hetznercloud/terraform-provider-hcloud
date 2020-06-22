package hcloud

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func init() {
	resource.AddTestSweepers("hcloud_load_balancer", &resource.Sweeper{
		Name: "hcloud_load_balancer",
		F: func(r string) error {
			var mErr error

			if err := testSweepLoadBalancers(r); err != nil {
				mErr = multierror.Append(mErr, err)
			}
			if err := testSweepCertificates(r); err != nil {
				mErr = multierror.Append(mErr, err)
			}
			return mErr
		},
	})
}

func TestAccHcloudLoadBalancer_Basic(t *testing.T) {
	var loadBalancer hcloud.LoadBalancer
	rInt := acctest.RandInt()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccHcloudPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccHcloudCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckLoadBalancerConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckLoadBalancerExists("hcloud_load_balancer.foobar", &loadBalancer),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer.foobar", "name", fmt.Sprintf("foo-load-balancer-%d", rInt)),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer.foobar", "location", "nbg1"),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer.foobar", "load_balancer_type", "lb11"),
				),
			},
			{
				Config: testAccHcloudCheckLoadBalancerConfig_Renamed(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckLoadBalancerExists("hcloud_load_balancer.foobar", &loadBalancer),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer.foobar", "name", fmt.Sprintf("foo-load-balancer-renamed-%d", rInt)),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer.foobar", "location", "nbg1"),
				),
			},
			{
				Config: testAccHcloudCheckLoadBalancerConfig_withTargets(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckLoadBalancerExists("hcloud_load_balancer.foobar_targets", &loadBalancer),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer.foobar_targets", "name", fmt.Sprintf("foo-load-balancer-targets-%d", rInt)),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer.foobar_targets", "location", "nbg1"),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer.foobar_targets", "target.#", "1"),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer.foobar_targets", "target.0.type", "server"),
				),
			},
			{
				Config: testAccHcloudCheckLoadBalancerConfig_withMultipleTargets(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckLoadBalancerExists("hcloud_load_balancer.foobar_targets", &loadBalancer),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer.foobar_targets", "name", fmt.Sprintf("foo-load-balancer-targets-%d", rInt)),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer.foobar_targets", "location", "nbg1"),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer.foobar_targets", "target.#", "2"),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer.foobar_targets", "target.0.type", "server"),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer.foobar_targets", "target.1.type", "server"),
				),
			},
			{
				Config: testAccHcloudCheckLoadBalancerConfig_withMultipleTargetsDeleteTarget(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckLoadBalancerExists("hcloud_load_balancer.foobar_targets", &loadBalancer),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer.foobar_targets", "name", fmt.Sprintf("foo-load-balancer-targets-%d", rInt)),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer.foobar_targets", "location", "nbg1"),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer.foobar_targets", "target.#", "1"),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer.foobar_targets", "target.0.type", "server"),
				),
			},
		},
	})
}

func testAccHcloudCheckLoadBalancerConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_load_balancer" "foobar" {
  name       = "foo-load-balancer-%d"
  load_balancer_type = "lb11"
  location   = "nbg1"
  algorithm {
    type = "round_robin"
  }
}
`, rInt)
}

func testAccHcloudCheckLoadBalancerConfig_Renamed(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_load_balancer" "foobar" {
  name       = "foo-load-balancer-renamed-%d"
  load_balancer_type = "lb11"
  location   = "nbg1"
  algorithm {
    type = "round_robin"
  }
}
`, rInt)
}

func testAccHcloudCheckLoadBalancerConfig_withTargets(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_server" "foobar" {
  name       = "foo-server-%d"
  server_type = "cx11"
  image = "ubuntu-18.04"
}
resource "hcloud_load_balancer" "foobar_targets" {
  name       = "foo-load-balancer-targets-%d"
  load_balancer_type = "lb11"
  location   = "nbg1"
  algorithm {
    type = "round_robin"
  }
  target {
    type = "server"
    server_id = "${hcloud_server.foobar.id}"
  }
}
`, rInt, rInt)
}

func testAccHcloudCheckLoadBalancerConfig_withMultipleTargets(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_server" "foobar" {
  name       = "foo-server-%d"
  server_type = "cx11"
  image = "ubuntu-18.04"
}

resource "hcloud_server" "foobar2" {
  name       = "foo-server2-%d"
  server_type = "cx11"
  image = "ubuntu-18.04"
}
resource "hcloud_load_balancer" "foobar_targets" {
  name       = "foo-load-balancer-targets-%d"
  load_balancer_type = "lb11"
  location   = "nbg1"
  algorithm {
    type = "round_robin"
  }
  target {
    type = "server"
    server_id = "${hcloud_server.foobar.id}"
  }
  target {
    type = "server"
    server_id = "${hcloud_server.foobar2.id}"
  }
}
`, rInt, rInt, rInt)
}

func testAccHcloudCheckLoadBalancerConfig_withMultipleTargetsDeleteTarget(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_server" "foobar" {
  name       = "foo-server-%d"
  server_type = "cx11"
  image = "ubuntu-18.04"
}

resource "hcloud_server" "foobar2" {
  name       = "foo-server2-%d"
  server_type = "cx11"
  image = "ubuntu-18.04"
}
resource "hcloud_load_balancer" "foobar_targets" {
  name       = "foo-load-balancer-targets-%d"
  load_balancer_type = "lb11"
  location   = "nbg1"
  algorithm {
    type = "round_robin"
  }

  target {
    type = "server"
    server_id = "${hcloud_server.foobar2.id}"
  }
}
`, rInt, rInt, rInt)
}

func testAccHcloudCheckLoadBalancerDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*hcloud.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "hcloud_load_balancer" {
			continue
		}

		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Load Balancer id is no int: %v", err)
		}
		var loadBalancer *hcloud.LoadBalancer
		loadBalancer, _, err = client.LoadBalancer.GetByID(context.Background(), id)
		if err != nil {
			return fmt.Errorf(
				"Error checking if Load Balancer (%s) is deleted: %v",
				rs.Primary.ID, err)
		}
		if loadBalancer != nil {
			return fmt.Errorf("Load Balancer (%s) has not been deleted", rs.Primary.ID)
		}
	}

	return nil
}

func testAccHcloudCheckLoadBalancerExists(n string, loadBalancer *hcloud.LoadBalancer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		client := testAccProvider.Meta().(*hcloud.Client)
		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return err
		}

		// Try to find the key
		foundLoadBalancer, _, err := client.LoadBalancer.GetByID(context.Background(), id)
		if err != nil {
			return err
		}

		if foundLoadBalancer == nil {
			return fmt.Errorf("Record not found")
		}

		*loadBalancer = *foundLoadBalancer
		return nil
	}
}

func testSweepLoadBalancers(region string) error {
	client, err := createClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	loadBalancers, err := client.LoadBalancer.All(ctx)
	if err != nil {
		return err
	}

	for _, loadBalancer := range loadBalancers {
		if _, err := client.LoadBalancer.Delete(ctx, loadBalancer); err != nil {
			return err
		}
	}
	testSweepServers(region)
	return nil
}
