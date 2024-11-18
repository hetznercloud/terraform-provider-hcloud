# Assign existing ipv4 only
resource "hcloud_server" "server_test" {
  //...
  public_net {
    ipv4_enabled = true
    ipv4 = hcloud_primary_ip.primary_ip_1.id
    ipv6_enabled = false
  }
  //...
}
# Link a managed ipv4 but autogenerate ipv6
resource "hcloud_server" "server_test" {
  //...
  public_net {
    ipv4_enabled = true
    ipv4 = hcloud_primary_ip.primary_ip_1.id
    ipv6_enabled = true
  }
  //...
}
# Assign & create auto-generated ipv4 & ipv6
resource "hcloud_server" "server_test" {
  //...
  public_net {
    ipv4_enabled = true
    ipv6_enabled = true
  }
  //...
}
