package e2etests

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/hetznercloud/hcloud-go/hcloud"
	tfhcloud "github.com/hetznercloud/terraform-provider-hcloud/hcloud"
)

var (
	// TestImage is the system image that is used in all tests
	TestImage = getEnv("TEST_IMAGE", "ubuntu-22.04")

	// TestImage is the system image ID that is used in all tests
	TestImageID = getEnv("TEST_IMAGE_ID", "67794396")

	// TestServerType is the default server type used in all tests
	TestServerType = getEnv("TEST_SERVER_TYPE", "cx11")

	// TestArchitecture is the default architecture used in all tests, should match the architecture of the TestServerType.
	TestArchitecture = getEnv("TEST_ARCHITECTURE", string(hcloud.ArchitectureX86))

	// TestLoadBalancerType is the default Load Balancer type used in all tests
	TestLoadBalancerType = "lb11"

	// TestDataCenter is the default datacenter where we execute our tests.
	TestDataCenter = getEnv("TEST_DATACENTER", "nbg1-dc3")

	// TestLocationName is the default location where we execute our tests.
	TestLocationName = getEnv("TEST_LOCATION", "nbg1")
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

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
