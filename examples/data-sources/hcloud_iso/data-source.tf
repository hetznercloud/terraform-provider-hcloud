data "hcloud_iso" "by_id" {
  id = "117577"
}

data "hcloud_iso" "by_name_x86" {
  name              = "nixos-minimal-24.11.712431.cbd8ec4de446-x86_64-linux.iso"
  with_architecture = "x86"
}

data "hcloud_iso" "by_name_prefix_arm" {
  name_prefix       = "nixos-minimal-24.11"
  with_architecture = "arm"
}

resource "hcloud_server" "main" {
  iso = data.hcloud_iso.by_name_x86.id
}
