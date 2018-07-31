package hcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"testing"
)

func init() {
	resource.AddTestSweepers("hcloud_rdns", &resource.Sweeper{
		Name: "hcloud_rdns",
		F:    testSweepRDNS,
	})
}
func TestAccHcloudReverseDNSCreateAndChange(t *testing.T) {

	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccHcloudPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccHcloudCheckFloatingIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckReverseDnsConfig_server(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hcloud_rdns.rdns_server", "dns_ptr", "example.com"),
				),
			},
		},
	})
}
func testAccHcloudCheckReverseDnsConfig_server(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_ssh_key" "floating_ip" {
  name       = "rdns-%d"
  public_key = "%s"
}
resource "hcloud_server" "floating_ip1" {
  name        = "rdns-1-%d"
  server_type = "cx11"
	image       = "debian-9"
  datacenter  = "fsn1-dc8"
  ssh_keys    = ["${hcloud_ssh_key.floating_ip.id}"]
}
resource "hcloud_rdns" "rdns_server" {
  server_id   = "${hcloud_server.floating_ip1.id}"
  ip_address  = "${hcloud_server.floating_ip1.ipv4_address}"
  dns_ptr     = "example.com"
}
`, rInt, testAccSSHPublicKey, rInt)
}
func testSweepRDNS(region string) error {
	testSweepFloatingIps(region)
	testSweepServers(region)
	return nil
}
