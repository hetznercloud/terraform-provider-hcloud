data "hcloud_primary_ips" "ip_2" {
  with_selector = "key=value"
}
