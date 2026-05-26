terraform import hcloud_rdns.example "$PREFIX-$ID-$IP"

# A Server with id 132022102 and ip 203.0.113.10
terraform import hcloud_rdns.server1 "s-132022102-203.0.113.10"

# A Primary IP with id 582026301 and ip 2001:db8::1
terraform import hcloud_rdns.primary_ip1 "p-582026301-2001:db8::1"

# A Floating IP with id 912300308 and ip 2001:db8::1
terraform import hcloud_rdns.floating_ip1 "p-912300308-2001:db8::1"

# A Load Balancer with id 747590326 and ip 203.0.113.25
terraform import hcloud_rdns.load_balancer1 "p-747590326-203.0.113.25"
