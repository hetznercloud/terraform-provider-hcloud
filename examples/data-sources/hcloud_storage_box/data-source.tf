data "hcloud_storage_box" "by_id" {
  id = 1333
}

data "hcloud_storage_box" "by_name" {
  name = "backups"
}

data "hcloud_storage_box" "by_label_selector" {
  with_selector = "env=production"
}
