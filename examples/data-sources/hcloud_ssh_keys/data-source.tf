data "hcloud_ssh_keys" "all" {}

data "hcloud_ssh_keys" "filter_with_selector" {
  with_selector = "foo=bar"
}

resource "hcloud_server" "main" {
  ssh_keys = data.hcloud_ssh_keys.all.ssh_keys.*.name
}
