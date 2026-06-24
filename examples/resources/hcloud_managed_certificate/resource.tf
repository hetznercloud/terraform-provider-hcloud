resource "hcloud_managed_certificate" "main" {
  lifecycle {
    # Important: prevents downtime during replacement, and ensures dependencies first
    # stop using the certificate before it is deleted.
    create_before_destroy = true
  }

  name         = "example"
  domain_names = ["example.com", "*.example.com"]
  labels = {
    key = "value"
  }
}
