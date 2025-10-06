resource "hcloud_zone" "example" {
  name = provider::hcloud::idna("exämple-🍪.com")
  mode = "primary"
}
