data "hcloud_server_types" "ds" {
}

resource "hcloud_server" "workers" {
  count = 3

  name        = "node1"
  image       = "debian-11"
  server_type = element(data.hcloud_server_types.ds.names, count.index)
}
