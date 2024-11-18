data "hcloud_datacenters" "ds" {
}

resource "hcloud_server" "workers" {
  count = 5

  name        = "node${count.index}"
  image       = "debian-11"
  server_type = "cx22"
  datacenter  = element(data.hcloud_datacenters.ds.datacenters, count.index).name
}
