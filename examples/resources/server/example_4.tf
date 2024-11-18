# Get image infos because we need the ID
data "hcloud_image" "packer_snapshot" {
  with_selector = "app=foobar"
  most_recent = true
}

# Create a new server from the snapshot
resource "hcloud_server" "from_snapshot" {
  name        = "from-snapshot"
  image       = data.hcloud_image.packer_snapshot.id
  server_type = "cx22"
  public_net {
    ipv4_enabled = true
    ipv6_enabled = true
  }
}
