variable "storage_box_id" {}

data "hcloud_storage_box_snapshots" "all" {
  storage_box = var.storage_box_id
}

data "hcloud_storage_box_snapshots" "by_label_selector" {
  storage_box   = var.storage_box_id
  with_selector = "env=production"
}
