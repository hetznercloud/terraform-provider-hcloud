resource "hcloud_zone_rrset" "example_dkim" {
  zone = hcloud_zone.example.name
  name = "mail._domainkey"
  type = "TXT"
  records = [
    { value = provider::hcloud::txt_record("k=rsa; p=MIIBITANBgkqhkiG9w0BAQEFAAOCAQ4AMIIBCQKCAQBmEPzUl5TqkAPe2AGLVRLl9jxer4mBnvHiRRpUY17l7dZQG8RlP/d9NXGUz6bLDTz5renHrpt/4cg3hDuqlvMuYNY5bmc6zOxmh+ydLr3gRtCwOvDXTYp2W3Ujob8V/Iy7E7F3NNfjT3vk3xBwdniOmlMKbmfOQ/AKmuuKjXBNOuRd797UPc8IiDPuuAi1UKQfdkkBZOoQBSnVnEvReoyIJnJSJVv1W7B459WXKR6ENaOdTYiPwTB+eMc0MLAPzBZopeM4kaBfNoxCqbvcJ4rDbC+jIqeARuIGpsQPx57kpyVLdmj07uCDzdAa/3FAPhm6iB6ih45e1RpCgfrasSORAgMBAAE=") },
  ]
}

resource "hcloud_zone_rrset" "example_spf" {
  zone = hcloud_zone.example.name
  name = "@"
  type = "TXT"
  records = [
    { value = provider::hcloud::txt_record("v=spf1 include:_spf.example.net ~all") },
  ]
}
