### Server creation with one linked primary ip (ipv4)
resource "hcloud_primary_ip" "primary_ip_1" {
  name          = "primary_ip_test"
  location      = "hel1"
  type          = "ipv4"
  assignee_type = "server"
  auto_delete   = true
  labels = {
    "hallo" : "welt"
  }
}

resource "hcloud_server" "server_test" {
  name        = "test-server"
  image       = "ubuntu-24.04"
  server_type = "cx23"
  location    = "hel1"
  labels = {
    "test" : "tessst1"
  }
  public_net {
    ipv4_enabled = true
    ipv4         = hcloud_primary_ip.primary_ip_1.id
    ipv6_enabled = false
  }
}
