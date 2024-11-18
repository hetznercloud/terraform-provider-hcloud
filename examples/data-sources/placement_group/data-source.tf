data "hcloud_placement_group" "sample_placement_group_1" {
  name = "sample-placement-group-1"
}

data "hcloud_placement_group" "sample_placement_group_2" {
  id = "4711"
}
