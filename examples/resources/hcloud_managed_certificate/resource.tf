resource "hcloud_managed_certificate" "managed_cert" {
  name         = "managed_cert"
  domain_names = ["*.example.com", "example.com"]
  labels = {
    label_1 = "value_1"
    label_2 = "value_2"
    # ...
  }
}
