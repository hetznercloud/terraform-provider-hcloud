package testsupport

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"os"
	"testing"

	"github.com/hetznercloud/hcloud-go/hcloud"
	tfhcloud "github.com/hetznercloud/terraform-provider-hcloud/hcloud"
)

// AccTestProviders returns all providers used during acceptance testing.
func AccTestProviders() map[string]*schema.Provider {
	return map[string]*schema.Provider {
		"hcloud": tfhcloud.Provider(),
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

// AccTestPreCheck checks if all conditions for an acceptance test are
// met.
func AccTestPreCheck(t *testing.T) func() {
	return func() {
		if v := os.Getenv("HCLOUD_TOKEN"); v == "" {
			t.Fatal("HCLOUD_TOKEN must be set for acceptance tests")
		}
	}
}
