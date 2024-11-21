data "hcloud_datacenter" "by_id" {
  id = 4
}

data "hcloud_datacenter" "by_name" {
  name = "fsn1-dc14"
}
