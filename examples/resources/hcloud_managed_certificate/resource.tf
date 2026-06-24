locals {
  domain_names = ["example.com", "*.example.com"]
}

resource "time_static" "main" {
  triggers = {
    # Generate a new certificate name each time domain names changes
    domain_names = join("\n", local.domain_names)
  }
}

resource "hcloud_managed_certificate" "main" {
  lifecycle {
    # Important: prevents downtime during replacement, and ensures dependencies first
    # stop using the certificate before it is deleted.
    create_before_destroy = true
  }

  // Important: prevents name uniqueness error during replacement.
  name         = "example-${time_static.main.rfc3339}"
  domain_names = local.domain_names
  labels = {
    key = "value"
  }
}
