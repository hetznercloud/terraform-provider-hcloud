package firewall

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/firewall"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func init() {
	resource.AddTestSweepers(firewall.ResourceType, &resource.Sweeper{
		Name:         firewall.ResourceType,
		Dependencies: []string{},
		F:            Sweep,
	})
}

// Sweep removes all firewalls from the Hetzner Cloud backend.
func Sweep(r string) error {
	client, err := testsupport.CreateClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	firewalls, err := client.Firewall.All(ctx)
	if err != nil {
		return err
	}

	for _, fw := range firewalls {
		if _, err := client.Firewall.Delete(ctx, fw); err != nil {
			return err
		}
	}

	return nil
}

// ByID returns a function that obtains a firewall by its ID.
func ByID(t *testing.T, firewall *hcloud.Firewall) func(*hcloud.Client, int) bool {
	return func(c *hcloud.Client, id int) bool {
		found, _, err := c.Firewall.GetByID(context.Background(), id)
		if err != nil {
			t.Fatalf("find firewall %d: %v", id, err)
		}
		if found == nil {
			return false
		}
		if firewall != nil {
			*firewall = *found
		}
		return true
	}
}

// DData defines the fields for the "testdata/d/hcloud_firewall" template.
type DData struct {
	testtemplate.DataCommon

	FirewallID    string
	FirewallName  string
	LabelSelector string
}

// TFID returns the data source identifier.
func (d *DData) TFID() string {
	return fmt.Sprintf("data.%s.%s", firewall.DataSourceType, d.RName())
}

// DDataList defines the fields for the "testdata/d/hcloud_firewalls" template.
type DDataList struct {
	testtemplate.DataCommon

	LabelSelector string
}

// TFID DDataList the data source identifier.
func (d *DDataList) TFID() string {
	return fmt.Sprintf("data.%s.%s", firewall.DataSourceListType, d.RName())
}

// RData defines the fields for the "testdata/r/hcloud_firewall"
// template.
type RData struct {
	testtemplate.DataCommon

	Name    string
	Rules   []RDataRule
	ApplyTo []RDataApplyTo
	Labels  map[string]string
}

// NewRData creates data for a new firewall resource.
func NewRData(t *testing.T, name string, rules []RDataRule, applyTo []RDataApplyTo) *RData {
	rInt := acctest.RandInt()
	r := &RData{
		Name:    name,
		Rules:   rules,
		ApplyTo: applyTo,
		Labels:  map[string]string{"key": strconv.Itoa(rInt)},
	}
	r.SetRName(name)
	return r
}

// TFID returns the resource identifier.
func (d *RData) TFID() string {
	return fmt.Sprintf("%s.%s", firewall.ResourceType, d.RName())
}

// RDataRule defines the fields for the "testdata/r/hcloud_firewall" template.
type RDataRule struct {
	Direction      string
	Port           string
	SourceIPs      []string
	DestinationIPs []string
	Protocol       string
	Description    string
}

type RDataApplyTo struct {
	Server        string
	LabelSelector string
}

// RDataAttachment defines the fields for the
// "testdata/r/hcloud_firewall_attachment" template.
//
// Fields ending in Ref are meant to contain a string referencing a Terraform
// value.
type RDataAttachment struct {
	testtemplate.DataCommon

	FirewallIDRef  string
	ServerIDRefs   []string
	LabelSelectors []string
}

// NewRDataAttachment creates a new RDataAttachment with the passed
// terraform resource name. It references a firewall using fwIDRef.
func NewRDataAttachment(resName, fwIDRef string) *RDataAttachment {
	d := RDataAttachment{FirewallIDRef: fwIDRef}
	d.SetRName(resName)
	return &d
}

// TFID returns the resource identifier.
func (d *RDataAttachment) TFID() string {
	return fmt.Sprintf("%s.%s", firewall.AttachmentResourceType, d.RName())
}
