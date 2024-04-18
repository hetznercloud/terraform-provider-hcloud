package testsupport

import (
	"fmt"
	"log"
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
		hcloud.WithApplication("hcloud-terraform", "testing"),
		hcloud.WithDebugWriter(log.Writer()),
	}
	return hcloud.NewClient(opts...), nil
}
