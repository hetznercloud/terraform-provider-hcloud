import {
  to = hcloud_zone_record.spf
  identity = {
    zone  = "example.com"
    name  = "@"
    type  = "TXT"
    value = provider::hcloud::txt_record("v=spf1 include:_spf.example.net ~all")
  }
}
