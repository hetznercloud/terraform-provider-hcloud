package hcloud

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func init() {
	resource.AddTestSweepers("hcloud_floating_ip", &resource.Sweeper{
		Name: "hcloud_floating_ip",
		F:    testSweepFloatingIps,
	})
}

func TestAccHcloudFloatingIP_AssignAndUpdateDescription(t *testing.T) {
	var floatingIP hcloud.FloatingIP
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccHcloudPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccHcloudCheckFloatingIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckFloatingIPConfig_server(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckFloatingIPExists("hcloud_floating_ip.floating_ip", &floatingIP),
					resource.TestCheckResourceAttr(
						"hcloud_floating_ip.floating_ip", "home_location", "fsn1"),
					resource.TestCheckResourceAttr(
						"hcloud_floating_ip.floating_ip", "description", "test"),
					resource.TestCheckResourceAttr(
						"hcloud_floating_ip.floating_ip", "name", fmt.Sprintf("floating-ip-%d", rInt)),
				),
			},
			{
				Config: testAccHcloudCheckFloatingIPConfig_updateDescription(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckFloatingIPExists("hcloud_floating_ip.floating_ip", &floatingIP),
					resource.TestCheckResourceAttr(
						"hcloud_floating_ip.floating_ip", "home_location", "fsn1"),
					resource.TestCheckResourceAttr(
						"hcloud_floating_ip.floating_ip", "description", "updated test"),
					resource.TestCheckResourceAttr(
						"hcloud_floating_ip.floating_ip", "name", fmt.Sprintf("floating-ip-%d", rInt)),
				),
			},
			{
				Config: testAccHcloudCheckFloatingIPConfig_updateName(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckFloatingIPExists("hcloud_floating_ip.floating_ip", &floatingIP),
					resource.TestCheckResourceAttr(
						"hcloud_floating_ip.floating_ip", "home_location", "fsn1"),
					resource.TestCheckResourceAttr(
						"hcloud_floating_ip.floating_ip", "description", "updated test"),
					resource.TestCheckResourceAttr(
						"hcloud_floating_ip.floating_ip", "name", fmt.Sprintf("floating-ip-updated-%d", rInt)),
				),
			},
		},
	})
}

func testAccHcloudCheckFloatingIPConfig_server(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_ssh_key" "floating_ip" {
  name       = "floating-ip-%d"
  public_key = "%s"
}
resource "hcloud_server" "floating_ip1" {
  name        = "floating-ip-1-%d"
  server_type = "cx11"
  image       = "debian-9"
  datacenter  = "fsn1-dc14"
  ssh_keys    = ["${hcloud_ssh_key.floating_ip.id}"]
}

resource "hcloud_floating_ip" "floating_ip" {
  server_id   = "${hcloud_server.floating_ip1.id}"
  type        = "ipv4"
  description = "test"
  name        = "floating-ip-%d"
}`, rInt, testAccSSHPublicKey, rInt, rInt)
}

func testAccHcloudCheckFloatingIPConfig_updateDescription(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_ssh_key" "floating_ip" {
  name       = "floating-ip-%d"
  public_key = "%s"
}
resource "hcloud_server" "floating_ip1" {
  name        = "floating-ip-1-%d"
  server_type = "cx11"
  image       = "debian-9"
  datacenter  = "fsn1-dc14"
  ssh_keys    = ["${hcloud_ssh_key.floating_ip.id}"]
}

resource "hcloud_floating_ip" "floating_ip" {
  server_id   = "${hcloud_server.floating_ip1.id}"
  type        = "ipv4"
  description = "updated test"
  name        = "floating-ip-%d"
}`, rInt, testAccSSHPublicKey, rInt, rInt)
}

func testAccHcloudCheckFloatingIPConfig_updateName(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_ssh_key" "floating_ip" {
  name       = "floating-ip-%d"
  public_key = "%s"
}
resource "hcloud_server" "floating_ip1" {
  name        = "floating-ip-1-%d"
  server_type = "cx11"
  image       = "debian-9"
  datacenter  = "fsn1-dc14"
  ssh_keys    = ["${hcloud_ssh_key.floating_ip.id}"]
}

resource "hcloud_floating_ip" "floating_ip" {
  server_id   = "${hcloud_server.floating_ip1.id}"
  type        = "ipv4"
  description = "updated test"
  name        = "floating-ip-updated-%d"
}`, rInt, testAccSSHPublicKey, rInt, rInt)
}

func testAccHcloudCheckFloatingIPDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*hcloud.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "hcloud_floating_ip" {
			continue
		}

		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Floating IP id is no int: %v", err)
		}
		var floatingIP *hcloud.FloatingIP
		floatingIP, _, err = client.FloatingIP.GetByID(context.Background(), id)
		if err != nil {
			return fmt.Errorf(
				"Error checking if floating ip (%s) is deleted: %v",
				rs.Primary.ID, err)
		}
		if floatingIP != nil {
			return fmt.Errorf("Floating ip (%s) has not been deleted", rs.Primary.ID)
		}
	}

	return nil
}

func testAccHcloudCheckFloatingIPExists(n string, floatingIP *hcloud.FloatingIP) resource.TestCheckFunc {
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
			return fmt.Errorf("Floating IP id is no int: %v", err)
		}
		foundFloatingIP, _, err := client.FloatingIP.GetByID(context.Background(), id)
		if err != nil {
			return err
		}
		if foundFloatingIP == nil {
			return fmt.Errorf("Floating IP not found")
		}

		*floatingIP = *foundFloatingIP
		return nil
	}
}

func testSweepFloatingIps(region string) error {
	client, err := createClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	ips, err := client.FloatingIP.All(ctx)
	if err != nil {
		return err
	}

	for _, ip := range ips {
		if _, err := client.FloatingIP.Delete(ctx, ip); err != nil {
			return err
		}
	}

	return nil
}
