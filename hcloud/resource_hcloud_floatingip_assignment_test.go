package hcloud

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func init() {
	resource.AddTestSweepers("hcloud_floating_ip_assignment", &resource.Sweeper{
		Name: "hcloud_floating_ip_assignment",
		F:    testSweepFloatingIps,
	})
}

func TestAccHcloudFloatingIPAssignment_Create(t *testing.T) {
	var server hcloud.Server
	var floatingIP hcloud.FloatingIP
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccHcloudPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccHcloudCheckFloatingIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckFloatingIPAssignmentConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckFloatingIPExists("hcloud_floating_ip.foobar", &floatingIP),
					testAccHcloudCheckFloatingIPAssignmentFloatingIP("hcloud_floating_ip_assignment.foobar", &floatingIP),
					testAccHcloudCheckServerExists("hcloud_server.foobar", &server),
					testAccHcloudCheckFloatingIPAssignmentServer("hcloud_floating_ip_assignment.foobar", &server),
				),
			},
		},
	})
}

func testAccHcloudCheckFloatingIPAssignmentFloatingIP(n string, floatingIP *hcloud.FloatingIP) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}
		id := rs.Primary.Attributes["floating_ip_id"]

		if id != strconv.Itoa(floatingIP.ID) {
			return fmt.Errorf("Floating IP Assignment Floating IP id is not valid: %v", id)
		}

		return nil
	}
}

func testAccHcloudCheckFloatingIPAssignmentServer(n string, server *hcloud.Server) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}
		id := rs.Primary.Attributes["server_id"]

		if id != strconv.Itoa(server.ID) {
			return fmt.Errorf("Floating IP Assignment Server id is not valid: %v", id)
		}

		return nil
	}
}

func testAccHcloudCheckFloatingIPAssignmentConfig(serverID int) string {
	return fmt.Sprintf(`
resource "hcloud_server" "foobar" {
  name        = "foo-%d"
  server_type = "cx11"
  image       = "debian-9"
  datacenter  = "fsn1-dc8"
}

resource "hcloud_floating_ip" "foobar" {
  type          = "ipv4"
  home_location = "nbg1"
}

resource "hcloud_floating_ip_assignment" "foobar" {
  floating_ip_id = "${hcloud_floating_ip.foobar.id}"
  server_id      = "${hcloud_server.foobar.id}"
}
`, serverID)
}
