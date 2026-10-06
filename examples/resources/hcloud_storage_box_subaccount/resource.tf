variable "team_badger_password" {
  type      = string
  sensitive = true
}

resource "hcloud_storage_box" "main" {
  // ...
}

resource "hcloud_storage_box_subaccount" "team_badger" {
  storage_box_id = hcloud_storage_box.main.id

  name           = "badger"
  home_directory = "teams/badger/"
  password       = var.team_badger_password

  access_settings = {
    # Needs to be accessible from everyones home network
    reachable_externally = true
    # The team attaches the storage as network drives
    samba_enabled = true
  }

  description = "Primary account for the Badger team to upload files."
  labels = {
    env  = "production"
    team = "badger"
  }
}

# Keep the password out of the state: an ephemeral value can only be assigned to
# a write-only argument, and `password_wo_version` is what a change is read off.

ephemeral "random_password" "team_beaver" {
  length = 32
}

resource "hcloud_storage_box_subaccount" "team_beaver" {
  storage_box_id = hcloud_storage_box.main.id

  name           = "beaver"
  home_directory = "teams/beaver/"

  password_wo         = ephemeral.random_password.team_beaver.result
  password_wo_version = 1
}
