resource "hcloud_load_balancer" "main" {
  name               = "main"
  load_balancer_type = "lb11"
  network_zone       = "eu-central"
}

resource "hcloud_network" "network" {
  name     = "network"
  ip_range = "10.0.0.0/16"
}

resource "hcloud_network_subnet" "subnet" {
  network_id   = hcloud_network.network.id
  type         = "cloud"
  network_zone = "eu-central"
  ip_range     = "10.0.1.0/24"
}

resource "hcloud_load_balancer_network" "attachment" {
  load_balancer_id = hcloud_load_balancer.main.id
  subnet_id        = hcloud_network_subnet.subnet.id
  ip               = "10.0.1.5"
}
