package loadbalancer_test

import (
	"github.com/hetznercloud/terraform-provider-hcloud/internal/loadbalancer"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
)

// LoadBalancerRData is a resource for use in load balancer related test.
func LoadBalancerRData() *loadbalancer.RData {
	return &loadbalancer.RData{
		Name:         "basic-load-balancer",
		LocationName: teste2e.TestLocationName,
	}
}
