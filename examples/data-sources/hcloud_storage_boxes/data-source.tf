data "hcloud_storage_boxes" "all" {
}

data "hcloud_storage_boxes" "by_label_selector" {
  with_selector = "env=production"
}
