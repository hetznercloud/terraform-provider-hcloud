resource "hcloud_server" "node1" {
  name        = "node1"
  image       = "debian-12"
  server_type = "cx23"
}

resource "hcloud_snapshot" "my-snapshot" {
  server_id = hcloud_server.node1.id
}
