data "hcloud_zones" "all" {}

data "hcloud_zones" "by_label" {
  with_selector = "key=value"
}
