import {
  id = "${hcloud_load_balancer.example.id}-${hcloud_network.example.id}"
  to = hcloud_load_balancer_network.attachment
}
