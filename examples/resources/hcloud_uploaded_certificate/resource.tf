resource "hcloud_uploaded_certificate" "main" {
  lifecycle {
    # Important: prevents downtime during replacement, and ensures dependencies first
    # stop using the certificate before it is deleted.
    create_before_destroy = true
  }

  name = "example"

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

  labels = {
    key = "value"
  }
}
