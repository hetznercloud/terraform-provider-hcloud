data "hcloud_ssh_key" "ssh_key_1" {
  id = "1234"
}
data "hcloud_ssh_key" "ssh_key_2" {
  name = "my-ssh-key"
}
data "hcloud_ssh_key" "ssh_key_3" {
  fingerprint = "43:51:43:a1:b5:fc:8b:b7:0a:3a:a9:b1:0f:66:73:a8"
}
data "hcloud_ssh_key" "ssh_key_4" {
  with_selector = "key=value"
}
resource "hcloud_server" "main" {
  ssh_keys = [data.hcloud_ssh_key.ssh_key_1.id, data.hcloud_ssh_key.ssh_key_2.id, data.hcloud_ssh_key.ssh_key_3.id]
}
