package network

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/testsupport"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/testtemplate"
)

func init() {
	resource.AddTestSweepers(ResourceType, &resource.Sweeper{
		Name: ResourceType,
		F:    Sweep,
	})
}

// Sweep removes all Networks from the Hetzner Cloud backend.
func Sweep(r string) error {
	client, err := testsupport.CreateClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	networks, err := client.Network.All(ctx)
	if err != nil {
		return err
	}

	for _, nw := range networks {
		if _, err := client.Network.Delete(ctx, nw); err != nil {
			return err
		}
	}

	return nil
}

// ByID returns a function that obtains a network by its ID.
func ByID(t *testing.T, nw *hcloud.Network) func(*hcloud.Client, int) bool {
	return func(c *hcloud.Client, id int) bool {
		found, _, err := c.Network.GetByID(context.Background(), id)
		if err != nil {
			t.Fatalf("network by ID: %d: %v", id, err)
		}
		if found == nil {
			return false
		}
		if nw != nil {
			*nw = *found
		}
		return true
	}
}

// RData defines the fields for the "testdata/r/hcloud_network" template.
type RData struct {
	testtemplate.DataCommon

	Name    string
	IPRange string
	Labels  map[string]string
}

// RDataSubnet defines the fields for the "testdata/r/hcloud_network_subnet"
// template.
type RDataSubnet struct {
	testtemplate.DataCommon

	Name        string
	Type        string
	NetworkID   string
	NetworkZone string
	IPRange     string
}
