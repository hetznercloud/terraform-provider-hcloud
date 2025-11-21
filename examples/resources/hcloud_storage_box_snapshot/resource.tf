resource "hcloud_storage_box" "main" {
  // ...
}

resource "hcloud_storage_box_snapshot" "backup" {
  storage_box = hcloud_storage_box.main.id

  description = "Before Tool XYZ Migration"
  labels = {
    env = "production"
  }
}
