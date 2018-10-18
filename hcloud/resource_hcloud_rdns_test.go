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

func TestAccHcloudReverseDNSServerCreateAndChange(t *testing.T) {
	var server hcloud.Server
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccHcloudPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckReverseDNSConfigServer(rInt, "example.hetzner.cloud"),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckServerExists("hcloud_server.rdns1", &server),
					// IPv4
					resource.TestCheckResourceAttr(
						"hcloud_rdns.rdns_server_v4", "dns_ptr", "example.hetzner.cloud"),

					// IPv6
					resource.TestCheckResourceAttr(
						"hcloud_rdns.rdns_server_v6", "dns_ptr", "example.hetzner.cloud"),
				),
			},
			{
				Config: testAccHcloudCheckReverseDNSConfigServer(rInt, "new-example.hetzner.cloud"),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckServerExists("hcloud_server.rdns1", &server),
					// IPv4
					resource.TestCheckResourceAttr(
						"hcloud_rdns.rdns_server_v4", "dns_ptr", "new-example.hetzner.cloud"),

					// IPv6
					resource.TestCheckResourceAttr(
						"hcloud_rdns.rdns_server_v6", "dns_ptr", "new-example.hetzner.cloud"),
				),
			},
		},
	})
}

func TestAccHcloudReverseDNSFloatingIpIpv4CreateAndChange(t *testing.T) {
	var floatingIP hcloud.FloatingIP
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccHcloudPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckReverseDNSIPv4ConfigFloatingIP(rInt, "example.hetzner.cloud"),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckFloatingIPExists("hcloud_floating_ip.floating_ip_v4", &floatingIP),
					resource.TestCheckResourceAttr(
						"hcloud_rdns.rdns_floating_ip_v4", "dns_ptr", "example.hetzner.cloud"),
				),
			},
			{
				Config: testAccHcloudCheckReverseDNSIPv4ConfigFloatingIP(rInt, "new-example.hetzner.cloud"),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckFloatingIPExists("hcloud_floating_ip.floating_ip_v4", &floatingIP),
					resource.TestCheckResourceAttr(
						"hcloud_rdns.rdns_floating_ip_v4", "dns_ptr", "new-example.hetzner.cloud"),
				),
			},
		},
	})
}

func TestAccHcloudReverseDNSFloatingIpIpv6CreateAndChange(t *testing.T) {
	var floatingIP hcloud.FloatingIP
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccHcloudPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckReverseDNSIPv6ConfigFloatingIP(rInt, "example.hetzner.cloud"),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckFloatingIPExists("hcloud_floating_ip.floating_ip_v6", &floatingIP),
					resource.TestCheckResourceAttr(
						"hcloud_rdns.rdns_floating_ip_v6", "dns_ptr", "example.hetzner.cloud"),
				),
			},
			{
				Config: testAccHcloudCheckReverseDNSIPv6ConfigFloatingIP(rInt, "new-example.hetzner.cloud"),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckFloatingIPExists("hcloud_floating_ip.floating_ip_v6", &floatingIP),
					resource.TestCheckResourceAttr(
						"hcloud_rdns.rdns_floating_ip_v6", "dns_ptr", "new-example.hetzner.cloud"),
				),
			},
		},
	})
}

func testAccHcloudCheckReverseDNSConfigServer(rInt int, dnsPtr string) string {
	return fmt.Sprintf(`
resource "hcloud_ssh_key" "foobar_rdns" {
  name       = "foobar-%d"
  public_key = "%s"
}
resource "hcloud_server" "rdns1" {
  name        = "rdns-ipv4-%d"
  server_type = "cx11"
  image       = "debian-9"
  datacenter  = "fsn1-dc14"
  ssh_keys    = ["${hcloud_ssh_key.foobar_rdns.id}"]
}
resource "hcloud_rdns" "rdns_server_v4" {
  server_id   = "${hcloud_server.rdns1.id}"
  ip_address  = "${hcloud_server.rdns1.ipv4_address}"
  dns_ptr     = "%s"
}
resource "hcloud_rdns" "rdns_server_v6" {
	server_id   = "${hcloud_server.rdns1.id}"
	ip_address  = "${hcloud_server.rdns1.ipv6_address}"
	dns_ptr     = "%s"
}
`, rInt, testAccRDNSSSHPublicKey, rInt, dnsPtr, dnsPtr)
}

func testAccHcloudCheckReverseDNSIPv4ConfigFloatingIP(rInt int, dnsPtr string) string {
	return fmt.Sprintf(`
resource "hcloud_ssh_key" "foobar_rdns" {
  name       = "foobar-%d"
  public_key = "%s"
}
resource "hcloud_floating_ip" "floating_ip_v4" {
  type      = "ipv4"
	home_location = "fsn1"
}
resource "hcloud_rdns" "rdns_floating_ip_v4" {
  floating_ip_id = "${hcloud_floating_ip.floating_ip_v4.id}"
  ip_address     = "${hcloud_floating_ip.floating_ip_v4.ip_address}"
  dns_ptr        = "%s"
}
`, rInt, testAccRDNSSSHPublicKey, dnsPtr)
}

func testAccHcloudCheckReverseDNSIPv6ConfigFloatingIP(rInt int, dnsPtr string) string {
	return fmt.Sprintf(`
resource "hcloud_ssh_key" "foobar_rdns" {
  name       = "foobar-%d"
  public_key = "%s"
}
resource "hcloud_floating_ip" "floating_ip_v6" {
  type      = "ipv6"
	home_location = "fsn1"
}
resource "hcloud_rdns" "rdns_floating_ip_v6" {
  floating_ip_id = "${hcloud_floating_ip.floating_ip_v6.id}"
  ip_address     = "${hcloud_floating_ip.floating_ip_v6.ip_address}"
  dns_ptr        = "%s"
}
`, rInt, testAccRDNSSSHPublicKey, dnsPtr)
}

func testSweepRDNS(region string) error {
	testSweepFloatingIps(region)
	testSweepServers(region)
	return nil
}
