data "hcloud_network" "network_1" {
  id = "1234"
}
data "hcloud_network" "network_2" {
  name = "my-network"
}
data "hcloud_network" "network_3" {
  with_selector = "key=value"
}
