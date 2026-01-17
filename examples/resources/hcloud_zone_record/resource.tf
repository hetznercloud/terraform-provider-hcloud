# repo-1/email.tf
resource "hcloud_zone_record" "spf" {
  zone    = "example.com"
  name    = "@"
  type    = "TXT"
  value   = provider::hcloud::txt_record("v=spf1 include:_spf.example.net ~all")
  comment = "For email provider XYZ"
}

# repo-2/google_site_verification.tf
resource "hcloud_zone_record" "site_verification" {
  zone  = "example.com"
  name  = "@"
  type  = "TXT"
  value = provider::hcloud::txt_record("google-site-verification=1234")
}
