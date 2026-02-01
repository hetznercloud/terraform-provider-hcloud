variable "storage_box_id" {}

data "hcloud_storage_box_subaccount" "by_id" {
  storage_box_id = var.storage_box_id
  id             = 2
}

data "hcloud_storage_box_subaccount" "by_name" {
  storage_box_id = var.storage_box_id
  name           = "badger"
}

data "hcloud_storage_box_subaccount" "by_username" {
  storage_box_id = var.storage_box_id
  username       = "u507137-sub1"
}

data "hcloud_storage_box_subaccount" "by_label_selector" {
  storage_box_id = var.storage_box_id
  with_selector  = "team=billing"
}
