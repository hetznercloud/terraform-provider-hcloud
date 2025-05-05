data "hcloud_image" "by_id" {
  id = "114690387"
}

data "hcloud_image" "by_name_x86" {
  name              = "debian-12"
  with_architecture = "x86"
}

data "hcloud_image" "by_name_arm" {
  name              = "debian-12"
  with_architecture = "arm"
}

data "hcloud_image" "by_label" {
  with_selector = "key=value"
}

resource "hcloud_server" "main" {
  image = data.hcloud_image.by_name.id
}
