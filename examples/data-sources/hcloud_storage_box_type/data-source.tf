data "hcloud_storage_box_type" "by_id" {
  id = 1333
}

data "hcloud_server_type" "by_name" {
  name = "bx11"
}
