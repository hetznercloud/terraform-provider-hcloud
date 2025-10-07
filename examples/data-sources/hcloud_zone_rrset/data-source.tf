data "hcloud_zone" "example" {
  name = "example.com"
}

data "hcloud_zone_rrset" "by_id" {
  zone = data.hcloud_zone.example.name
  id   = "www/A"
}

data "hcloud_zone_rrset" "by_name_and_type" {
  zone = data.hcloud_zone.example.name
  name = "www"
  type = "A"
}

data "hcloud_zone_rrset" "by_label" {
  zone          = data.hcloud_zone.example.name
  with_selector = "key=value"
}
