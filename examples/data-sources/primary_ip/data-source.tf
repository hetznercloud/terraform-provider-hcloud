data "hcloud_primary_ip" "ip_1" {
  ip_address = "1.2.3.4"
}
data "hcloud_primary_ip" "ip_2" {
  name = "primary_ip_1"
}
data "hcloud_primary_ip" "ip_3" {
  with_selector = "key=value"
}

// Link a server to an existing primary IP
resource "hcloud_server" "server_test" {
  name        = "test-server"
  image       = "ubuntu-20.04"
  server_type = "cx22"
  datacenter  = "fsn1-dc14"
  labels = {
    "test" : "tessst1"
  }
  public_net {
    ipv4 = hcloud_primary_ip.ip_1.id
  }

}
