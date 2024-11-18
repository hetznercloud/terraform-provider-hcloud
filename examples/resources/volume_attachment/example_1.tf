resource "hcloud_volume_attachment" "main" {
  volume_id = hcloud_volume.master.id
  server_id = hcloud_server.node1.id
  automount = true
}

resource "hcloud_server" "node1" {
  name        = "node1"
  image       = "debian-11"
  server_type = "cx22"
  datacenter  = "nbg1-dc3"
}

resource "hcloud_volume" "master" {
  location = "nbg1"
  size     = 10
}
