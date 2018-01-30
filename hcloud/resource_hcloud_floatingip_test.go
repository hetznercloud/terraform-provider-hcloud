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

func TestAccFloatingIP_Server(t *testing.T) {
	var floatingIP hcloud.FloatingIP
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFloatingIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckFloatingIPConfig_server(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFloatingIPExists("hcloud_floating_ip.foobar", &floatingIP),
					resource.TestCheckResourceAttr(
						"hcloud_floating_ip.foobar", "home_location", "fsn1"),
				),
			},
		},
	})
}

func testAccCheckFloatingIPConfig_server(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_ssh_key" "foobar" {
  name       = "foobar-%d"
  public_key = "%s"
}
resource "hcloud_server" "foobar" {
  name        = "foo-%d"
  server_type = "cx11"
	image       = "debian-9"
  datacenter  = "fsn1-dc8"
  ssh_keys    = ["${hcloud_ssh_key.foobar.id}"]
}

resource "hcloud_floating_ip" "foobar" {
  server_id = "${hcloud_server.foobar.id}"
  type      = "ipv4"
}`, rInt, testAccSSHPublicKey, rInt)
}

func testAccCheckFloatingIPDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*hcloud.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "hcloud_floating_ip" {
			continue
		}

		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Floating IP id is no int: %v", err)
		}
		_, _, err = client.FloatingIP.GetByID(context.Background(), id)
		if err != nil {
			return fmt.Errorf(
				"Error waiting for floating ip (%s) to be destroyed: %v",
				rs.Primary.ID, err)
		}
	}

	return nil
}

func testAccCheckFloatingIPExists(n string, floatingIP *hcloud.FloatingIP) resource.TestCheckFunc {
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
