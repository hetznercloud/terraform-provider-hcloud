resource "hcloud_server" "node1" {
  name        = "node1"
  image       = "debian-11"
  server_type = "cx23"
}

resource "hcloud_volume" "master" {
  name      = "volume1"
  size      = 50
  server_id = hcloud_server.node1.id
  automount = true
  format    = "ext4"
}
