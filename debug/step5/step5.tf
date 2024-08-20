resource "hcloud_primary_ip" "primary-ip-v6" {
  name          = "primaryip-v6-test--3230555060468491167"
  type          = "ipv6"
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
    ipv4_enabled = false
    ipv6         = hcloud_primary_ip.primary-ip-v6.id
    ipv6_enabled = true
  }
}
