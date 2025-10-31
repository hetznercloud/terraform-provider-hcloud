resource "hcloud_server" "node1" {
  name        = "node1"
  image       = "debian-12"
  server_type = "cx23"
}

resource "hcloud_network" "network" {
  name     = "network"
  ip_range = "10.0.0.0/16"
}

resource "hcloud_network_subnet" "subnet1" {
  network_id   = hcloud_network.network.id
  type         = "cloud"
  network_zone = "eu-central"
  ip_range     = "10.0.1.0/24"
}

resource "hcloud_server_network" "node1_subnet1" {
  server_id = hcloud_server.node1.id
  subnet_id = hcloud_network_subnet.subnet1.id
  ip        = "10.0.1.5"
  alias_ips = [
    "10.0.1.10"
  ]
}
