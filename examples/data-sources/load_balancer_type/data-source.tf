data "hcloud_load_balancer_type" "by_name" {
  name = "cx22"
}

data "hcloud_load_balancer_type" "by_id" {
  id = 1
}

resource "hcloud_load_balancer" "load_balancer" {
  name               = "my-load-balancer"
  load_balancer_type = data.hcloud_load_balancer_type.name
  location           = "nbg1"
}
