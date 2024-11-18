resource "hcloud_load_balancer" "load_balancer1" {
  name               = "load_balancer1"
  load_balancer_type = "lb11"
  location           = "fsn1"
}

resource "hcloud_rdns" "load_balancer_master" {
  load_balancer_id = "${hcloud_load_balancer.load_balancer1.id}"
  ip_address       = "${hcloud_load_balancer.load_balancer1.ipv4}"
  dns_ptr          = "example.com"
}
