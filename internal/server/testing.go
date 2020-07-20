package server

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
		Name:         ResourceType,
		Dependencies: []string{},
		F:            Sweep,
	})
}

// Sweep removes all Servers from the Hetzner Cloud backend.
func Sweep(r string) error {
	client, err := testsupport.CreateClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	servers, err := client.Server.All(ctx)
	if err != nil {
		return err
	}

	for _, srv := range servers {
		if _, err := client.Server.Delete(ctx, srv); err != nil {
			return err
		}
	}

	return nil
}

// ByID returns a function that obtains a server by its ID.
func ByID(t *testing.T, srv *hcloud.Server) func(*hcloud.Client, int) bool {
	return func(c *hcloud.Client, id int) bool {
		found, _, err := c.Server.GetByID(context.Background(), id)
		if err != nil {
			t.Fatalf("find server %d: %v", id, err)
		}
		if found == nil {
			return false
		}
		if srv != nil {
			*srv = *found
		}
		return true
	}
}

// RData defines the fields for the "testdata/r/hcloud_server" template.
type RData struct {
	testtemplate.DataCommon

	Name         string
	Type         string
	Image        string
	LocationName string
	DataCenter   string
	SSHKeys      []string
	KeepDisk     bool
	Rescue       bool
	Backups      bool
	Labels       map[string]string
	UserData     string
}

// RDataNetwork defines the fields for the "testdata/r/hcloud_server_network"
// template.
type RDataNetwork struct {
	testtemplate.DataCommon

	Name      string
	ServerID  string
	NetworkID string
	IP        string
	AliasIPs  []string
}
