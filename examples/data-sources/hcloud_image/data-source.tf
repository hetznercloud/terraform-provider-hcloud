data "hcloud_image" "image_1" {
  id = "1234"
}
data "hcloud_image" "image_2" {
  name              = "ubuntu-18.04"
  with_architecture = "x86"
}
data "hcloud_image" "image_3" {
  with_selector = "key=value"
}

resource "hcloud_server" "main" {
  image = data.hcloud_image.image_1.id
}
