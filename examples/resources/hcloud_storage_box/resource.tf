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

resource "hcloud_storage_box" "ssh_key" {
  name             = "backups"
  storage_box_type = "bx21"
  location         = "hel1"
  password         = var.storage_box_password

  # You can set the initial SSH Keys as an attribute on the resource, but these
  # can not be updated through the API and through the terraform provider.
  # If this attribute is ever changed, the provider will mark the resource as
  # "requires replacement" and you could loose the data stored on the Storage Box.
  ssh_keys = [
    hcloud_ssh_key.my_key.public_key,
    file("~/.ssh/id_ed25519.pub"),
  ]

  lifecycle {
    # To avoid accidentally deleting the resource you can ignore changes in the ssh_keys attribute:
    ignore_changes = [
      ssh_keys
    ]

    # If you want to be extra sure that your Storage Box is not accidentally deleted:
    prevent_destroy = true
  }
}
