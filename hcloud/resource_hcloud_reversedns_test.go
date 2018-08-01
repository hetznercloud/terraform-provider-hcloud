package hcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"testing"
)

func init() {
	resource.AddTestSweepers("hcloud_rdns", &resource.Sweeper{
		Name: "hcloud_rdns",
		F:    testSweepRDNS,
	})
}
func TestAccHcloudReverseDNSCreateAndChange(t *testing.T) {
	var server hcloud.Server
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccHcloudPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testacchcloudcheckreversednsconfigServer(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckServerExists("hcloud_server.rdns1", &server),
					resource.TestCheckResourceAttr(
						"hcloud_rdns.rdns_server", "dns_ptr", "example-updated.com"),
				),
			},
			{
				Config: testacchcloudcheckreversednsconfigFloatingIp(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckServerExists("hcloud_server.rdns1", &server),
					resource.TestCheckResourceAttr(
						"hcloud_rdns.rdns_floating_ip", "dns_ptr", "floating-ip-updated.com"),
				),
			},
		},
	})
}
func testacchcloudcheckreversednsconfigServer(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_ssh_key" "rdns" {
  name       = "rdns-%d"
  public_key = "%s"
}
resource "hcloud_server" "rdns1" {
  name        = "rdns-1-%d"
  server_type = "cx11"
  image       = "debian-9"
  datacenter  = "fsn1-dc8"
  ssh_keys    = ["${hcloud_ssh_key.rdns.id}"]
}
resource "hcloud_rdns" "rdns_server" {
  server_id   = "${hcloud_server.rdns1.id}"
  ip_address  = "${hcloud_server.rdns1.ipv4_address}"
  dns_ptr     = "example.com"
}
resource "hcloud_rdns" "rdns_server" {
  server_id   = "${hcloud_server.rdns1.id}"
  ip_address  = "${hcloud_server.rdns1.ipv4_address}"
  dns_ptr     = "example-updated.com"
}
`, rInt, testAccSSHPublicKey, rInt)
}
func testacchcloudcheckreversednsconfigFloatingIp(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_ssh_key" "rdns" {
  name       = "rdns-%d"
  public_key = "%s"
}
resource "hcloud_server" "rdns2" {
  name        = "rdns-1-%d"
  server_type = "cx11"
  image       = "debian-9"
  datacenter  = "fsn1-dc8"
  ssh_keys    = ["${hcloud_ssh_key.rdns.id}"]
}
resource "hcloud_floating_ip" "rdns_floating_ip" {
  type      = "ipv4"
  server_id = "${hcloud_server.rdns2.id}"
}

resource "hcloud_rdns" "rdns_floating_ip" {
  floating_ip_id = "${hcloud_floating_ip.rdns_floating_ip.id}"
  ip_address     = "${hcloud_floating_ip.rdns_floating_ip.ip_address}"
  dns_ptr        = "floating-ip.com"
}
resource "hcloud_rdns" "rdns_floating_ip" {
  floating_ip_id = "${hcloud_floating_ip.rdns_floating_ip.id}"
  ip_address     = "${hcloud_floating_ip.rdns_floating_ip.ip_address}"
  dns_ptr        = "floating-ip-updated.com"
}
`, rInt, testAccSSHPublicKey, rInt)
}
func testSweepRDNS(region string) error {
	testSweepFloatingIps(region)
	testSweepServers(region)
	return nil
}
