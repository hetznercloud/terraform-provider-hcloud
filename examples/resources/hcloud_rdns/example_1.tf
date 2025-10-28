resource "hcloud_server" "node1" {
  name        = "node1"
  image       = "debian-12"
  server_type = "cx23"
}

resource "hcloud_rdns" "master" {
  server_id  = hcloud_server.node1.id
  ip_address = hcloud_server.node1.ipv4_address
  dns_ptr    = "example.com"
}
