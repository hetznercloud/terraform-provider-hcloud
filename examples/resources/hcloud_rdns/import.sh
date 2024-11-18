terraform import hcloud_rdns.example "$PREFIX-$ID-$IP"

# import reverse dns entry on server with id 123, ip 192.168.100.1
terraform import hcloud_rdns.myrdns s-123-192.168.100.1

# import reverse dns entry on primary ip with id 123, ip 2001:db8::1
terraform import hcloud_rdns.myrdns p-123-2001:db8::1

# import reverse dns entry on floating ip with id 123, ip 2001:db8::1
terraform import hcloud_rdns.myrdns f-123-2001:db8::1

# import reverse dns entry on load balancer with id 123, ip 2001:db8::1
terraform import hcloud_rdns.myrdns l-123-2001:db8::1
