data "hcloud_load_balancer_type" "by_id" {
  id = 1
}

data "hcloud_load_balancer_type" "by_name" {
  name = "lb11"
}

resource "hcloud_load_balancer" "main" {
  name               = "my-load-balancer"
  load_balancer_type = data.hcloud_load_balancer_type.name
  location           = "fsn1"
}
