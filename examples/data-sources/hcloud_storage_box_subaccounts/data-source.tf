variable "storage_box_id" {}

data "hcloud_storage_box_subaccounts" "all" {
  storage_box = var.storage_box_id
}

data "hcloud_storage_box_subaccounts" "by_label_selector" {
  storage_box   = var.storage_box_id
  with_selector = "team=billing"
}
