data "hcloud_volume" "volume_1" {
  id = "1234"
}
data "hcloud_volume" "volume_2" {
  name = "my-volume"
}
data "hcloud_volume" "volume_3" {
  with_selector = "key=value"
}
