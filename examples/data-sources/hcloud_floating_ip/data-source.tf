data "hcloud_floating_ip" "ip_1" {
  ip_address = "1.2.3.4"
}
data "hcloud_floating_ip" "ip_2" {
  with_selector = "key=value"
}
resource "hcloud_floating_ip_assignment" "main" {
  count          = var.counter
  floating_ip_id = data.hcloud_floating_ip.ip_1.id
  server_id      = hcloud_server.main.id
}
