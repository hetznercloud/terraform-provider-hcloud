resource "hcloud_floating_ip" "floating1" {
  home_location = "nbg1"
  type          = "ipv4"
}

resource "hcloud_rdns" "floating_master" {
  floating_ip_id = "${hcloud_floating_ip.floating1.id}"
  ip_address     = "${hcloud_floating_ip.floating1.ip_address}"
  dns_ptr        = "example.com"
}
