resource "hcloud_server" "test_server" {
    name        = "test-server"
    server_type = "cx22"
    image       = "ubuntu-20.04"
}

resource "hcloud_firewall" "basic_firewall" {
    name   = "basic_firewall"
}

resource "hcloud_firewall_attachment" "fw_ref" {
    firewall_id = hcloud_firewall.basic_firewall.id
    server_ids  = [hcloud_server.test_server.id]
}
