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
