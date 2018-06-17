package hcloud

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func init() {
	resource.AddTestSweepers("hcloud_floating_ip_association", &resource.Sweeper{
		Name: "hcloud_floating_ip_association",
		F:    testSweepFloatingIps,
	})
}

func TestAccHcloudFloatingIPAssociation_Create(t *testing.T) {
	floatingIP := hcloud.FloatingIP{ID: acctest.RandInt()}
	server := hcloud.Server{ID: acctest.RandInt()}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccHcloudPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccHcloudCheckFloatingIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckFloatingIPAssociationConfig(server.ID),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckFloatingIPExists("hcloud_floating_ip.foobar", &floatingIP),
					resource.TestCheckResourceAttr(
						"hcloud_floating_ip_association.foobar", "floating_ip_id", strconv.Itoa(floatingIP.ID)),
					testAccHcloudCheckServerExists("hcloud_server.foobar", &server),
					resource.TestCheckResourceAttr(
						"hcloud_floating_ip_association.foobar", "server_id", strconv.Itoa(server.ID))),
			},
		},
	})
}

func testAccHcloudCheckFloatingIPAssociationConfig(serverID int) string {
	return fmt.Sprintf(`
resource "hcloud_server" "foobar" {
  name        = "foo-%d"
  server_type = "cx11"
  image       = "debian-9"
  datacenter  = "fsn1-dc8"
  ssh_keys    = ["${hcloud_ssh_key.foobar.id}"]
}

resource "hcloud_floating_ip" "foobar" {
  type          = "ipv4"
  home_location = "nbg1"
}

resource "hcloud_floating_ip_association" "foobar" {
  type           = "ipv4"
  floating_ip_id = "${hcloud_floating_ip.foobar.id}"
  server_id      = "${hcloud_server.foobar.id}"
}
`, serverID)
}
