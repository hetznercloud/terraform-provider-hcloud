resource "hcloud_network" "mynet" {
  name     = "my-net"
  ip_range = "10.0.0.0/8"
}
resource "hcloud_network_route" "privNet" {
  network_id  = hcloud_network.mynet.id
  destination = "10.100.1.0/24"
  gateway     = "10.0.1.1"
}
