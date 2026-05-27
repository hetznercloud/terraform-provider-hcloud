import {
  to = hcloud_rdns.example
  id = "$RESOURCE_PREFIX-$RESOURCE_ID-$IP_ADDRESS"

  # A Server with id 132022102 and ip 203.0.113.10
  # id = "s-132022102-203.0.113.10"

  # A Primary IP with id 582026301 and ip 2001:db8::1
  # id = "p-582026301-2001:db8::1"

  # A Floating IP with id 912300308 and ip 2001:db8::1
  # id = "f-912300308-2001:db8::1"

  # A Load Balancer with id 747590326 and ip 203.0.113.25
  # id = "l-747590326-203.0.113.25"
}
