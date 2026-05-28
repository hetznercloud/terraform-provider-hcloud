// For Servers
resource "hcloud_server" "server1" {
  name = "server1"
  // ...
}

resource "hcloud_rdns" "server1" {
  server_id  = hcloud_server.server1.id
  ip_address = hcloud_server.server1.ipv4_address
  dns_ptr    = "example.com"
}

// For Primary IPs
resource "hcloud_primary_ip" "primary_ip1" {
  name = "primary_ip1"
  type = "ipv4"
  // ...
}

resource "hcloud_rdns" "primary_ip1" {
  primary_ip_id = hcloud_primary_ip.primary_ip1.id
  ip_address    = hcloud_primary_ip.primary_ip1.ip_address
  dns_ptr       = "example.com"
}

// For Floating IPs
resource "hcloud_floating_ip" "floating_ip1" {
  name = "floating_ip1"
  type = "ipv4"
  // ...
}

resource "hcloud_rdns" "floating_ip1" {
  floating_ip_id = hcloud_floating_ip.floating_ip1.id
  ip_address     = hcloud_floating_ip.floating_ip1.ip_address
  dns_ptr        = "example.com"
}

// For Load Balancers
resource "hcloud_load_balancer" "load_balancer1" {
  name = "load_balancer1"
  // ...
}

resource "hcloud_rdns" "load_balancer1" {
  load_balancer_id = hcloud_load_balancer.load_balancer1.id
  ip_address       = hcloud_load_balancer.load_balancer1.ipv4
  dns_ptr          = "example.com"
}
