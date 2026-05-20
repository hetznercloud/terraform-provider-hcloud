resource "hcloud_server" "example" {
  name        = "example"
  server_type = "cx22"
  image       = "debian-12"
}

resource "hcloud_server_power_state" "example" {
  server_id = hcloud_server.example.id
  state     = "off"
}
