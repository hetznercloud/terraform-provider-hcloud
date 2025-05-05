data "hcloud_images" "by_architecture" {
  with_architecture = ["x86"]
}

data "hcloud_images" "by_label" {
  with_selector = "key=value"
}
