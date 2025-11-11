resource "hcloud_storage_box" "backups" {
  name             = "backups"
  storage_box_type = "bx21"
  location         = "hel1"
  password         = var.storage_box_password

  labels = {
    "foo" : "bar"
  }

  access_settings = {
    reachable_externally = true
    samba_enabled        = true
    ssh_enabled          = true
    webdav_enabled       = true
    zfs_enabled          = true
  }

  snapshot_plan = {
    max_snapshots = 10
    minute        = 16
    hour          = 18
    day_of_week   = 3
  }

  delete_protection = true
}
