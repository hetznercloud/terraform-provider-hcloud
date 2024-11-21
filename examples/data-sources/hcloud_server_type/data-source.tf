data "hcloud_server_type" "by_id" {
  id = 22
}

data "hcloud_server_type" "by_name" {
  name = "cx22"
}

resource "hcloud_server" "main" {
  name        = "my-server"
  location    = "fsn1"
  image       = "debian-12"
  server_type = data.hcloud_server_type.by_name.name
}
