resource "hcloud_primary_ip" "primary1" {
  datacenter = "nbg1-dc3"
  type       = "ipv4"
}

resource "hcloud_rdns" "primary1" {
  primary_ip_id = hcloud_primary_ip.primary1.id
  ip_address    = hcloud_primary_ip.primary1.ip_address
  dns_ptr       = "example.com"
}
