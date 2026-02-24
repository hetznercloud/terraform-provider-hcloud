package server

import (
	"context"
	"fmt"
	"testing"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

// GetAPIResource returns a [testsupport.GetAPIResourceFunc] for [hcloud.Server].
func GetAPIResource() testsupport.GetAPIResourceFunc[hcloud.Server] {
	return func(c *hcloud.Client, attrs map[string]string) (*hcloud.Server, error) {
		result, _, err := c.Server.Get(context.Background(), attrs["id"])
		return result, err
	}
}

// ByID returns a function that obtains a server by its ID.
func ByID(t *testing.T, srv *hcloud.Server) func(*hcloud.Client, int64) bool {
	return func(c *hcloud.Client, id int64) bool {
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

// DData defines the fields for the "testdata/d/hcloud_server"
// template.
type DData struct {
	testtemplate.DataCommon

	ServerID      string
	ServerName    string
	LabelSelector string
}

// TFID returns the data source identifier.
func (d *DData) TFID() string {
	return fmt.Sprintf("data.%s.%s", DataSourceType, d.RName())
}

// DDataList defines the fields for the "testdata/d/hcloud_servers" template.
type DDataList struct {
	testtemplate.DataCommon

	LabelSelector string
}

// TFID DDataList the data source identifier.
func (d *DDataList) TFID() string {
	return fmt.Sprintf("data.%s.%s", DataSourceListType, d.RName())
}

// RData defines the fields for the "testdata/r/hcloud_server" template.
type RData struct {
	testtemplate.DataCommon

	Name                   string
	Type                   string
	Image                  string
	LocationName           string
	Datacenter             string
	PublicNet              map[string]interface{}
	SSHKeys                []string
	KeepDisk               bool
	Rescue                 string
	Backups                bool
	ISO                    string
	Labels                 map[string]string
	UserData               string
	Networks               []RDataInlineNetwork
	FirewallIDs            []string
	DependsOn              []string
	PlacementGroupID       string
	DeleteProtection       bool
	RebuildProtection      bool
	AllowDeprecatedImages  bool
	ShutdownBeforeDeletion bool

	Raw string
}

// RDataInlineNetwork defines the information required to attach a server
// to a network directly in the server resource.
type RDataInlineNetwork struct {
	NetworkID string
	IP        string
	AliasIPs  []string
}

// TFID returns the resource identifier.
func (d *RData) TFID() string {
	return fmt.Sprintf("%s.%s", ResourceType, d.RName())
}

// RDataNetwork defines the fields for the "testdata/r/hcloud_server_network"
// template.
type RDataNetwork struct {
	testtemplate.DataCommon

	Name      string
	ServerID  string
	NetworkID string
	SubNetID  string
	IP        string
	AliasIPs  []string
	DependsOn []string
}

// TFID returns the resource identifier.
func (d *RDataNetwork) TFID() string {
	return fmt.Sprintf("%s.%s", NetworkResourceType, d.RName())
}

// AData defines the fields for the "testdata/a/hcloud_server"
// template.
type AData struct {
	testtemplate.DataCommon

	Type     string
	ServerID string
}

// TFID returns the resource identifier.
func (d *AData) TFID() string {
	return fmt.Sprintf("action.hcloud_server_%s.%s", d.Type, d.RName())
}
