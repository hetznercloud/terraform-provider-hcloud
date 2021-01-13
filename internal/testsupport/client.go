package testsupport

import (
	"fmt"
	"os"

	"github.com/hetznercloud/hcloud-go/hcloud"
)

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
