resource "hcloud_zone" "example_primary" {
  name = "example.com"
  mode = "primary"

  ttl = 10800

  labels = {
    key = "value"
  }

  delete_protection = false
}


resource "hcloud_zone" "example_secondary" {
  name = "example.com"
  mode = "secondary"

  ttl = 10800

  labels = {
    key = "value"
  }

  primary_nameservers = [
    {
      address = "203.5.113.54"
    },
    {
      address = "203.5.113.55"
      port    = 5353
    },
    {
      address  = "203.5.113.56"
      tsig_alg = "hmac-sha256"
      tsig_key = "205ab4910f35e361cefd99727082d6d08f4597f715164704d6541b3bc33da98f"
    },
  ]

  delete_protection = false
}

resource "hcloud_zone" "example_idna" {
  name = provider::hcloud::idna("exämple-🍪.com")
  mode = "primary"
}
