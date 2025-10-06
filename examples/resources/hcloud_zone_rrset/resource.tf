resource "hcloud_zone" "example" {
  name = "example.com"
  mode = "primary"
}

resource "hcloud_zone_rrset" "example" {
  zone = hcloud_zone.example.name
  name = "www"
  type = "A"

  ttl = 10800

  labels = {
    key = "value"
  }

  records = [
    { value = "201.78.10.45", comment = "web server 1" },
  ]

  change_protection = false
}
