data "hcloud_ssh_keys" "all_keys" {
}
data "hcloud_ssh_keys" "keys_by_selector" {
  with_selector = "foo=bar"
}
resource "hcloud_server" "main" {
  ssh_keys = data.hcloud_ssh_keys.all_keys.ssh_keys.*.name
}
