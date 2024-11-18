data "hcloud_locations" "ds" {
}

resource "hcloud_server" "workers" {
  count = 5

  name        = "node${count.index}"
  image       = "debian-11"
  server_type = "cx22"
  location    = element(data.hcloud_locations.ds.locations, count.index).name
}
