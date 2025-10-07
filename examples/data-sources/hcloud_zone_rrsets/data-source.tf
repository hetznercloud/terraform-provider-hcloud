data "hcloud_zone" "example" {
  name = "example.com"
}

data "hcloud_zone_rrsets" "all" {
  zone = data.hcloud_zone.example.name
}

data "hcloud_zone_rrsets" "by_label" {
  zone          = data.hcloud_zone.example.name
  with_selector = "key=value"
}
