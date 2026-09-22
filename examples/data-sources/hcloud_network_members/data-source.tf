data "hcloud_network_members" "by_type" {
  network_id = hcloud_network.example.id
  with_type  = ["load_balancer", "server"]
}

data "hcloud_network_members" "by_status" {
  network_id  = hcloud_network.example.id
  with_status = ["ok", "attaching", "detaching", "updating", "error"]
}

data "hcloud_network_members" "by_subnet" {
  network_id  = hcloud_network.example.id
  with_subnet = ["10.0.1.0/24"]
}

data "hcloud_network_members" "all" {
  network_id = hcloud_network.example.id
}

locals {
  // A list of all IPs in the network
  network_ips = flatten([
    for m in data.hcloud_network_members.all.members :
    concat([m.ip], tolist(m.alias_ips))
  ])
}
