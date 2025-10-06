data "hcloud_zone" "by_id" {
  id = 1234
}

data "hcloud_zone" "by_name" {
  name = "example.com"
}

data "hcloud_zone" "by_label" {
  with_selector = "key=value"
}
