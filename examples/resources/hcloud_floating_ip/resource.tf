resource "hcloud_server" "node1" {
  name        = "node1"
  image       = "debian-12"
  server_type = "cx23"
}

resource "hcloud_floating_ip" "master" {
  type      = "ipv4"
  server_id = hcloud_server.node1.id
}
