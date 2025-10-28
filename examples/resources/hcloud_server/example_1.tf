# Create a new server running debian
resource "hcloud_server" "node1" {
  name        = "node1"
  image       = "debian-12"
  server_type = "cx23"
  public_net {
    ipv4_enabled = true
    ipv6_enabled = true
  }
}
