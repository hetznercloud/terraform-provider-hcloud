resource "hcloud_network" "privNet" {
  name     = "my-net"
  ip_range = "10.0.1.0/24"
}
