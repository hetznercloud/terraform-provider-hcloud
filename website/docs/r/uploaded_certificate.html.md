---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_uploaded_certificate"
sidebar_current: "docs-hcloud-resource-uploaded-certificate-x"
description: |-
  Upload a TLS certificate to Hetzner Cloud.
---

# hcloud_uploaded_certificate

Upload a TLS certificate to Hetzner Cloud.

## Example Usage

```hcl
resource "hcloud_uploaded_certificate" "sample_certificate" {
    name = "test-certificate-%d"

    private_key =<<-EOT
    -----BEGIN RSA PRIVATE KEY-----
    MIIEpQIBAAKCAQEAorPccsHibgGLJIub5Sb1yvDvARifoKzg7MIhyAYLnJkGn9B1
    ...
    AHcjLFCNvobInLHTTmCoAxYBmEv2eakas0+n4g/LM2Ukaw1Bz+3VrVo=
    -----END RSA PRIVATE KEY-----
    EOT

    certificate =<<-EOT
    -----BEGIN CERTIFICATE-----
    MIIDMDCCAhigAwIBAgIIJgROscP8RRUwDQYJKoZIhvcNAQELBQAwIDEeMBwGA1UE
    ...
    TKS8gQ==
    -----END CERTIFICATE-----
    EOT

    labels = {
        label_1 = "value_1"
        label_2 = "value_2"
        ...
    }
}
```

## Argument Reference

- `name` - (Required, string) Name of the Certificate.
- `private_key` - (Required, string) PEM encoded private key belonging to the certificate.
- `certificate` - (Required, string) PEM encoded TLS certificate.
- `labels` - (Optional, map) User-defined labels (key-value pairs) the
certificate should be created with.

## Attribute Reference

- `id` - (int) Unique ID of the certificate.
- `name` - (string) Name of the Certificate.
- `certificate` - (string) PEM encoded TLS certificate.
- `labels` - (map) User-defined labels (key-value pairs) assigned to the certificate.
- `domain_names` - (list) Domains and subdomains covered by the certificate.
- `fingerprint` - (string) Fingerprint of the certificate.
- `created` - (string) Point in time when the Certificate was created at Hetzner Cloud (in ISO-8601 format).
- `not_valid_before` - (string) Point in time when the Certificate becomes valid (in ISO-8601 format).
- `not_valid_after` - (string) Point in time when the Certificate stops being valid (in ISO-8601 format).

## Import

Uploaded certificates can be imported using their `id`:

```hcl
terraform import hcloud_uploaded_certificate.sample_certificate id
```
