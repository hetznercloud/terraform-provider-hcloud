data "hcloud_load_balancer" "lb_1" {
  name = "my-load-balancer"
}
data "hcloud_load_balancer" "lb_2" {
  id = "123"
}
data "hcloud_load_balancer" "lb_3" {
  with_selector = "key=value"
}
