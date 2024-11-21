data "hcloud_ssh_key" "by_id" {
  id = 24332897
}

data "hcloud_ssh_key" "by_name" {
  name = "my-ssh-key"
}

data "hcloud_ssh_key" "by_fingerprint" {
  fingerprint = "55:58:dc:bd:61:6e:7d:24:07:a7:7d:9b:be:99:83:a8"
}

data "hcloud_ssh_key" "by_label" {
  with_selector = "key=value"
}

resource "hcloud_server" "main" {
  ssh_keys = [
    data.hcloud_ssh_key.by_id.id,
    data.hcloud_ssh_key.by_name.id,
    data.hcloud_ssh_key.by_fingerprint.id,
  ]
}
