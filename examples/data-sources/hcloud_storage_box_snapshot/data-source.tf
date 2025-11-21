variable "storage_box_id" {}

data "hcloud_storage_box_snapshot" "by_id" {
  storage_box = var.storage_box_id
  id          = 2
}

data "hcloud_storage_box_snapshot" "by_name" {
  storage_box = var.storage_box_id
  name        = "2025-02-12T11-35-19"
}

data "hcloud_storage_box" "by_label_selector" {
  storage_box   = var.storage_box_id
  with_selector = "env=production"
}
