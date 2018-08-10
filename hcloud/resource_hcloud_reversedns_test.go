package hcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

var testAccRDNSSSHPublicKey string

func init() {
	resource.AddTestSweepers("hcloud_rdns", &resource.Sweeper{
		Name: "hcloud_rdns",
		F:    testSweepRDNS,
	})

	var err error
	testAccRDNSSSHPublicKey, _, err = acctest.RandSSHKeyPair("hcloud@ssh-acceptance-test")
	if err != nil {
		panic(fmt.Errorf("Cannot generate test SSH key pair: %s", err))
	}
}

func TestAccHcloudReverseDNSServerIPv4CreateAndChange(t *testing.T) {
	var server hcloud.Server
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccHcloudPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckReverseDNSIPv4ConfigServer(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckServerExists("hcloud_server.rdns1", &server),
					resource.TestCheckResourceAttr(
						"hcloud_rdns.rdns_server_v4", "dns_ptr", "example.com"),
				),
			},
		},
	})
}
func TestAccHcloudReverseDNSServerIPv6CreateAndChange(t *testing.T) {
	var server hcloud.Server
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccHcloudPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckReverseDNSIPv6ConfigServer(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckServerExists("hcloud_server.rdns4", &server),
					resource.TestCheckResourceAttr(
						"hcloud_rdns.rdns_server_v6", "dns_ptr", "example.com"),
				),
			},
		},
	})
}

func TestAccHcloudReverseDNSFloatingIpIpv4CreateAndChange(t *testing.T) {
	var server hcloud.Server
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccHcloudPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckReverseDNSIPv4ConfigFloatingIP(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckServerExists("hcloud_server.rdns2", &server),
					resource.TestCheckResourceAttr(
						"hcloud_rdns.rdns_floating_ip_v4", "dns_ptr", "floating-ip.com"),
				),
			},
		},
	})
}
func TestAccHcloudReverseDNSFloatingIpIpv6CreateAndChange(t *testing.T) {
	var server hcloud.Server
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccHcloudPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckReverseDNSIPv6ConfigFloatingIP(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckServerExists("hcloud_server.rdns3", &server),
					resource.TestCheckResourceAttr(
						"hcloud_rdns.floating_ip_v6", "dns_ptr", "floating-ip.com"),
				),
			},
		},
	})
}
func testAccHcloudCheckReverseDNSIPv4ConfigServer(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_ssh_key" "foobar_rdns" {
  name       = "foobar-%d"
  public_key = "%s"
}
resource "hcloud_server" "rdns1" {
  name        = "rdns-ipv4-%d"
  server_type = "cx11"
  image       = "debian-9"
  datacenter  = "fsn1-dc8"
  ssh_keys    = ["${hcloud_ssh_key.foobar_rdns.id}"]
}
resource "hcloud_rdns" "rdns_server_v4" {
  server_id   = "${hcloud_server.rdns1.id}"
  ip_address  = "${hcloud_server.rdns1.ipv4_address}"
  dns_ptr     = "example.com"
}
`, rInt, testAccRDNSSSHPublicKey, rInt)
}
func testAccHcloudCheckReverseDNSIPv6ConfigServer(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_ssh_key" "foobar_rdns" {
  name       = "foobar-%d"
  public_key = "%s"
}
resource "hcloud_server" "rdns2" {
  name        = "rdns-ipv6-%d"
  server_type = "cx11"
  image       = "debian-9"
  datacenter  = "fsn1-dc8"
  ssh_keys    = ["${hcloud_ssh_key.foobar_rdns.id}"]
}
resource "hcloud_rdns" "rdns_server_v6" {
  server_id   = "${hcloud_server.rdns2.id}"
  ip_address  = "${cidrhost(hcloud_server.rdns2.ipv6_network,2)}"
  dns_ptr     = "example.com"
}
`, rInt, testAccRDNSSSHPublicKey, rInt)
}
func testAccHcloudCheckReverseDNSIPv4ConfigFloatingIP(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_ssh_key" "foobar_rdns" {
  name       = "foobar-%d"
  public_key = "%s"
}
resource "hcloud_server" "rdns2" {
  name        = "rdns-ipv4-%d"
  server_type = "cx11"
  image       = "debian-9"
  datacenter  = "fsn1-dc8"
  ssh_keys    = ["${hcloud_ssh_key.foobar_rdns.id}"]
}
resource "hcloud_floating_ip" "floating_ip_v4" {
  type      = "ipv4"
  server_id = "${hcloud_server.rdns2.id}"
}

resource "hcloud_rdns" "rdns_floating_ip_v4" {
  floating_ip_id = "${hcloud_floating_ip.floating_ip_v4.id}"
  ip_address     = "${hcloud_floating_ip.floating_ip_v4.ip_address}"
  dns_ptr        = "floating-ip.com"
}
`, rInt, testAccRDNSSSHPublicKey, rInt)
}
func testAccHcloudCheckReverseDNSIPv6ConfigFloatingIP(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_ssh_key" "foobar_rdns" {
  name       = "foobar-%d"
  public_key = "%s"
}
resource "hcloud_server" "rdns3" {
  name        = "rdns-ipv6-%d"
  server_type = "cx11"
  image       = "debian-9"
  datacenter  = "fsn1-dc8"
  ssh_keys    = ["${hcloud_ssh_key.foobar_rdns.id}"]
}
resource "hcloud_floating_ip" "floating_ip_v6" {
  type      = "ipv6"
  server_id = "${hcloud_server.rdns3.id}"
}

resource "hcloud_rdns" "rdns_floating_ip_v6" {
  floating_ip_id = "${hcloud_floating_ip.floating_ip_v6.id}"
  ip_address     = "${cidrhost(hcloud_floating_ip.floating_ip_v6.ip_network,2)}"
  dns_ptr        = "floating-ip.com"
}
`, rInt, testAccRDNSSSHPublicKey, rInt)
}
func testSweepRDNS(region string) error {
	testSweepFloatingIps(region)
	testSweepServers(region)
	return nil
}
