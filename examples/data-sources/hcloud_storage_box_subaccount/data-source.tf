variable "storage_box_id" {}

data "hcloud_storage_box_subaccount" "by_id" {
  storage_box = var.storage_box_id
  id          = 2
}

data "hcloud_storage_box_subaccount" "by_username" {
  storage_box = var.storage_box_id
  username    = "2025-02-12T11-35-19"
}

data "hcloud_storage_box_subaccount" "by_label_selector" {
  storage_box   = var.storage_box_id
  with_selector = "team=billing"
}
