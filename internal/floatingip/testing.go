package floatingip

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/testsupport"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/testtemplate"
)

func init() {
	resource.AddTestSweepers(ResourceType, &resource.Sweeper{
		Name:         ResourceType,
		Dependencies: []string{},
		F:            Sweep,
	})
}

// Sweep removes all Floating IPs from the Hetzner Cloud backend.
func Sweep(r string) error {
	client, err := testsupport.CreateClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	servers, err := client.FloatingIP.All(ctx)
	if err != nil {
		return err
	}

	for _, srv := range servers {
		if _, err := client.FloatingIP.Delete(ctx, srv); err != nil {
			return err
		}
	}

	return nil
}

// ByID returns a function that obtains a Floating IP by its ID.
func ByID(t *testing.T, fl *hcloud.FloatingIP) func(*hcloud.Client, int) bool {
	return func(c *hcloud.Client, id int) bool {
		found, _, err := c.FloatingIP.GetByID(context.Background(), id)
		if err != nil {
			t.Fatalf("find floating ip %d: %v", id, err)
		}
		if found == nil {
			return false
		}
		if fl != nil {
			*fl = *found
		}
		return true
	}
}

// RData defines the fields for the "testdata/r/hcloud_floating_ip" template.
type RData struct {
	testtemplate.DataCommon

	Name             string
	Type             string
	HomeLocationName string
	Labels           map[string]string
}

// TFID returns the resource identifier.
func (d *RData) TFID() string {
	return fmt.Sprintf("%s.%s", ResourceType, d.RName())
}

// RDataAssignment defines the fields for the "testdata/r/hcloud_floating_ip_assignment" template.
type RDataAssignment struct {
	testtemplate.DataCommon

	FloatingIPID string
	ServerID     string
}

// TFID returns the resource identifier.
func (d *RDataAssignment) TFID() string {
	return fmt.Sprintf("%s.%s", AssignmentResourceType, d.RName())
}
