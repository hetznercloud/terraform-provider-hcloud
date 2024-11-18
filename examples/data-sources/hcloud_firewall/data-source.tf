data "hcloud_firewall" "sample_firewall_1" {
  name = "sample-firewall-1"
}

data "hcloud_firewall" "sample_firewall_2" {
  id = "4711"
}
