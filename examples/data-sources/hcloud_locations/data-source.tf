data "hcloud_locations" "all" {}

resource "hcloud_server" "workers" {
  count = 5

  name        = "node${count.index}"
  image       = "debian-12"
  server_type = "cx23"
  location    = element(data.hcloud_locations.all.locations, count.index).name
}
