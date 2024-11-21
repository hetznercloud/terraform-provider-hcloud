data "hcloud_datacenters" "all" {}

resource "hcloud_server" "workers" {
  count = 5

  name        = "node${count.index}"
  image       = "debian-12"
  server_type = "cx22"
  datacenter  = element(data.hcloud_datacenters.all.datacenters, count.index).name
}
