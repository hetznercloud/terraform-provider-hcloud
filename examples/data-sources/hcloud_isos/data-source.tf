data "hcloud_isos" "by_architecture" {
  with_architecture = ["x86"]
}

data "hcloud_isos" "by_name_prefix" {
  name_prefix = "nixos"
}

data "hcloud_isos" "by_type" {
  type = "private"
}
