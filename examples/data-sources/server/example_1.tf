data "hcloud_server" "s_1" {
  name = "my-server"
}
data "hcloud_server" "s_2" {
  id = "123"
}
data "hcloud_server" "s_3" {
  with_selector = "key=value"
}
