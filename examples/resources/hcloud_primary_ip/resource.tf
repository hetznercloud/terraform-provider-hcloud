resource "hcloud_primary_ip" "main" {
  name          = "primary_ip_test"
  datacenter    = "fsn1-dc14"
  type          = "ipv4"
  assignee_type = "server"
  auto_delete   = true
  labels = {
    "hallo" : "welt"
  }
}
// Link a server to a primary IP
resource "hcloud_server" "server_test" {
  name        = "test-server"
  image       = "ubuntu-24.04"
  server_type = "cx23"
  datacenter  = "fsn1-dc14"
  labels = {
    "test" : "tessst1"
  }
  public_net {
    ipv4 = hcloud_primary_ip.main.id
  }

}
