package e2etests

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
	tfhcloud "github.com/hetznercloud/terraform-provider-hcloud/hcloud"
)

const (
	// TestImage is the system image that is used in all tests
	TestImage = "ubuntu-20.04"

	// TestServerType is the default server type used in all tests
	TestServerType = "cx11"

	// TestLoadBalancerType is the default Load Balancer type used in all tests
	TestLoadBalancerType = "lb11"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

// Providers returns all providers used during acceptance testing.
func Providers() map[string]*schema.Provider {
	return map[string]*schema.Provider{
		"hcloud": tfhcloud.Provider(),
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

// CreateClient creates a new *hcloud.Client which authenticates using the
// HCLOUD_TOKEN variable.
func CreateClient() (*hcloud.Client, error) {
	if os.Getenv("HCLOUD_TOKEN") == "" {
		return nil, fmt.Errorf("empty HCLOUD_TOKEN")
	}
	opts := []hcloud.ClientOption{
		hcloud.WithToken(os.Getenv("HCLOUD_TOKEN")),
	}
	return hcloud.NewClient(opts...), nil
}
