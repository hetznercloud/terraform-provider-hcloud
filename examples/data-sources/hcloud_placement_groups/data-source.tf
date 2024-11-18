data "hcloud_placement_groups" "sample_placement_group_1" {

}

data "hcloud_placement_groups" "sample_placement_group_2" {
  with_selector = "key=value"
}
