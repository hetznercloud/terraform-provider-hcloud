resource "hcloud_primary_ip" "primary-ip-v4" {
  name          = "primaryip-v4-test--3230555060468491167"
  type          = "ipv4"
  datacenter    = "nbg1-dc3"
  assignee_type = "server"
  auto_delete   = false
}

resource "hcloud_server" "server-primaryIP-network-test" {
  name        = "server-primaryIP-network-test--3230555060468491167"
  server_type = "cpx11"
  image       = "ubuntu-24.04"
  datacenter  = "nbg1-dc3"

  public_net {
    ipv4         = hcloud_primary_ip.primary-ip-v4.id
    ipv4_enabled = true
    ipv6_enabled = true
  }
}
