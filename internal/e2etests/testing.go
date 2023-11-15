package e2etests

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/hetznercloud/hcloud-go/hcloud"
	tfhcloud "github.com/hetznercloud/terraform-provider-hcloud/hcloud"
)

const (
	// TestImage is the system image that is used in all tests
	TestImage = "ubuntu-20.04"

	// TestServerType is the default server type used in all tests
	TestServerType = "cx11"

	// TestArchitecture is the default architecture used in all tests, should match the architecture of the TestServerType.
	TestArchitecture = hcloud.ArchitectureX86

	// TestLoadBalancerType is the default Load Balancer type used in all tests
	TestLoadBalancerType = "lb11"

	// TestDataCenter is the default datacenter where we execute our tests.
	TestDataCenter = "hel1-dc2"

	// TestLocationName is the default location where we execute our tests.
	TestLocationName = "hel1"
)

func ProtoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"hcloud": func() (tfprotov6.ProviderServer, error) {
			ctx := context.Background()

			providerFactory, err := tfhcloud.GetMuxedProvider(ctx)
			if err != nil {
				return nil, err
			}

			return providerFactory(), nil
		},
	}
}

// PreCheck checks if all conditions for an acceptance test are
// met.
func PreCheck(t *testing.T) func() {
	return func() {
		if v := os.Getenv("HCLOUD_TOKEN"); v == "" {
			t.Fatal("HCLOUD_TOKEN must be set for acceptance tests")
		}
	}
}
