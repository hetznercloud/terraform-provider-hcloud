resource "hcloud_floating_ip_assignment" "main" {
  floating_ip_id = hcloud_floating_ip.master.id
  server_id      = hcloud_server.node1.id
}

resource "hcloud_server" "node1" {
  name        = "node1"
  image       = "debian-11"
  server_type = "cx23"
  datacenter  = "fsn1-dc8"
}

resource "hcloud_floating_ip" "master" {
  type          = "ipv4"
  home_location = "nbg1"
}
