data "hcloud_certificate" "sample_certificate_1" {
  name = "sample-certificate-1"
}

data "hcloud_certificate" "sample_certificate_2" {
  id = "4711"
}
