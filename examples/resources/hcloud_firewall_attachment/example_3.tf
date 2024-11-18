terraform {
  required_providers {
    hcloud = {
      source  = "hetznercloud/hcloud"
      version = "1.32.2"
    }
  }
}

resource "hcloud_firewall" "deny_all" {
  name = "deny_all"
}

resource "hcloud_server" "test_server" {
  name                       = "test-server"
  server_type                = "cx22"
  image                      = "ubuntu-20.04"
  ignore_remote_firewall_ids = true
  firewall_ids = [
    hcloud_firewall.deny_all.id
  ]
}

resource "hcloud_firewall" "allow_rules" {
  name = "allow_rules"

  rule {
    direction = "in"
    protocol  = "tcp"
    port      = "22"
    source_ips = [
      "0.0.0.0/0",
      "::/0",
    ]
    destination_ips = [
      format("%s/32", hcloud_server.test_server.ipv4_address)
    ]
  }
}

resource "hcloud_firewall_attachment" "deny_all_att" {
  firewall_id = hcloud_firewall.deny_all.id
  server_ids  = [hcloud_server.test_server.id]
}

resource "hcloud_firewall_attachment" "allow_rules_att" {
  firewall_id = hcloud_firewall.allow_rules.id
  server_ids  = [hcloud_server.test_server.id]
}
