resource "hcloud_placement_group" "my-placement-group" {
  name = "my-placement-group"
  type = "spread"
  labels = {
    key = "value"
  }
}

resource "hcloud_server" "node1" {
  name               = "node1"
  image              = "debian-11"
  server_type        = "cx23"
  placement_group_id = hcloud_placement_group.my-placement-group.id
}
