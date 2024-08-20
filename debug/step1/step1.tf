resource "hcloud_server" "server-primaryIP-network-test" {
  name        = "server-primaryIP-network-test--3230555060468491167"
  server_type = "cpx11"
  image       = "ubuntu-24.04"
  datacenter  = "nbg1-dc3"

  public_net {
    ipv4_enabled = true
    ipv6_enabled = true
  }
}
