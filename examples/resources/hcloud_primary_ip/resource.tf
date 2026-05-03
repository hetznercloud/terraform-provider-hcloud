resource "hcloud_primary_ip" "main" {
  name        = "primary-ip"
  location    = "fsn1"
  type        = "ipv4"
  auto_delete = false

  labels = {
    key = "value"
  }
}

// Link a server to a primary IP
resource "hcloud_server" "main" {
  name        = "server"
  image       = "ubuntu-24.04"
  server_type = "cx23"
  location    = "fsn1"

  public_net {
    ipv4 = hcloud_primary_ip.main.id
  }
}
