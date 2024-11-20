data "hcloud_images" "image_2" {
  with_architecture = ["x86"]
}

data "hcloud_images" "image_3" {
  with_selector = "key=value"
}
