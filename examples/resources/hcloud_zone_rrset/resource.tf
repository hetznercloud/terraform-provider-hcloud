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

resource "hcloud_zone_rrset" "example_soa" {
  zone = hcloud_zone.example.name
  name = "@"
  type = "SOA"

  records = [
    // SOA record SERIAL value will be set to 0, before saving it to the state. Make
    // sure to use 0 as SERIAL value to prevent running into inconsistent state errors.
    { value = "hydrogen.ns.hetzner.com. dns.hetzner.com. 0 86400 10800 3600000 3600" }
    //                                                   ^ here
  ]
}
