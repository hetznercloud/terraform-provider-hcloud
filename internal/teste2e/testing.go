package teste2e

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/joho/godotenv"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	tfhcloud "github.com/hetznercloud/terraform-provider-hcloud/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
)

var (
	// TestImage is the system image that is used in all tests
	TestImage = getEnv("TEST_IMAGE", "ubuntu-24.04")

	// TestImage is the system image ID that is used in all tests
	TestImageID = getEnv("TEST_IMAGE_ID", "161547269")

	// TestServerType is the default server type used in all tests
	TestServerType = getEnv("TEST_SERVER_TYPE", "cpx22")

	// TestServerTypeUpgrade is the upgrade server type used in all tests
	TestServerTypeUpgrade = getEnv("TEST_SERVER_TYPE_UPGRADE", "cpx32")

	// TestArchitecture is the default architecture used in all tests, should match the architecture of the TestServerType.
	TestArchitecture = getEnv("TEST_ARCHITECTURE", string(hcloud.ArchitectureX86))

	// TestLoadBalancerType is the default Load Balancer type used in all tests
	TestLoadBalancerType = "lb11"

	// TestDataCenter is the default datacenter where we execute our tests.
	TestDataCenter = getEnv("TEST_DATACENTER", "hel1-dc2")

	// TestLocationName is the default location where we execute our tests.
	TestLocationName = getEnv("TEST_LOCATION", "hel1")
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
	// We load the environment variables file before the test case starts (for example
	// to load the TF_ACC=1 variable need by the acceptance tests)
	if err := loadEnvFile(filepath.Join(testsupport.ProjectRoot(t), ".env")); err != nil {
		t.Fatalf("Could not load .env file: %v", err)
	}

	return func() {
		if v := os.Getenv("HCLOUD_TOKEN"); v == "" {
			t.Fatal("HCLOUD_TOKEN must be set for acceptance tests")
		}
	}
}

func loadEnvFile(filename string) error {
	if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
		return nil
	}

	return godotenv.Load(filename)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
