locals {
  private_key = <<-EOT
    -----BEGIN RSA PRIVATE KEY-----
    MIIEpQIBAAKCAQEAorPccsHibgGLJIub5Sb1yvDvARifoKzg7MIhyAYLnJkGn9B1
    ...
    AHcjLFCNvobInLHTTmCoAxYBmEv2eakas0+n4g/LM2Ukaw1Bz+3VrVo=
    -----END RSA PRIVATE KEY-----
    EOT

  certificate = <<-EOT
    -----BEGIN CERTIFICATE-----
    MIIDMDCCAhigAwIBAgIIJgROscP8RRUwDQYJKoZIhvcNAQELBQAwIDEeMBwGA1UE
    ...
    TKS8gQ==
    -----END CERTIFICATE-----
    EOT
}

resource "time_static" "main" {
  triggers = {
    # Generate a new certificate name each time private key or certificate changes
    private_key = local.private_key
    certificate = local.certificate
  }
}

resource "hcloud_uploaded_certificate" "main" {
  lifecycle {
    # Important: prevents downtime during replacement, and ensures dependencies first
    # stop using the certificate before it is deleted.
    create_before_destroy = true
  }

  // Important: prevents name uniqueness error during replacement.
  name = "example-${time_static.main.rfc3339}"

  private_key = local.private_key
  certificate = local.certificate

  labels = {
    key = "value"
  }
}
