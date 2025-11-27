variable "storage_box_id" {}

data "hcloud_storage_box_subaccount" "by_id" {
  storage_box_id = var.storage_box_id
  id             = 2
}

data "hcloud_storage_box_subaccount" "by_username" {
  storage_box_id = var.storage_box_id
  username       = "billing-team"
}

data "hcloud_storage_box_subaccount" "by_label_selector" {
  storage_box_id = var.storage_box_id
  with_selector  = "team=billing"
}
