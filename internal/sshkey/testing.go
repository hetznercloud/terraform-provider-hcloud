package sshkey

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func init() {
	resource.AddTestSweepers(ResourceType, &resource.Sweeper{
		Name:         ResourceType,
		Dependencies: []string{},
		F:            Sweep,
	})
}

// Sweep removes all sshkeys from the Hetzner Cloud backend.
func Sweep(r string) error {
	client, err := testsupport.CreateClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	sshkeys, err := client.SSHKey.All(ctx)
	if err != nil {
		return err
	}

	for _, cert := range sshkeys {
		if _, err := client.SSHKey.Delete(ctx, cert); err != nil {
			return err
		}
	}

	return nil
}

// ByID returns a function that obtains a sshkey by its ID.
func ByID(t *testing.T, cert *hcloud.SSHKey) func(*hcloud.Client, int) bool {
	return func(c *hcloud.Client, id int) bool {
		found, _, err := c.SSHKey.GetByID(context.Background(), id)
		if err != nil {
			t.Fatalf("find sshkey %d: %v", id, err)
		}
		if found == nil {
			return false
		}
		if cert != nil {
			*cert = *found
		}
		return true
	}
}

// DData defines the fields for the "testdata/d/hcloud_ssh_key"
// template.
type DData struct {
	testtemplate.DataCommon

	SSHKeyID      string
	SSHKeyName    string
	LabelSelector string
}

// TFID returns the data source identifier.
func (d *DData) TFID() string {
	return fmt.Sprintf("data.%s.%s", DataSourceType, d.RName())
}

// SSHKeysDData defines the fields for the "testdata/d/hcloud_ssh_keys"
// template.
type SSHKeysDData struct {
	testtemplate.DataCommon
	LabelSelector string
}

// TFID returns the data source identifier.
func (d *SSHKeysDData) TFID() string {
	return fmt.Sprintf("data.%s.%s", SSHKeysDataSourceType, d.RName())
}

// RData defines the fields for the "testdata/r/hcloud_ssh_key"
// template.
type RData struct {
	testtemplate.DataCommon

	Name      string
	PublicKey string
	Labels    map[string]string
}

// TFID returns the resource identifier.
func (d *RData) TFID() string {
	return fmt.Sprintf("%s.%s", ResourceType, d.RName())
}

// NewRData creates data for a new sshkey resource.
func NewRData(t *testing.T, name string) *RData {
	publicKeyMaterial, _, err := acctest.RandSSHKeyPair("hcloud@ssh-acceptance-test")
	rInt := acctest.RandInt()
	if err != nil {
		t.Fatal(err)
	}
	r := &RData{
		Name:      name,
		PublicKey: publicKeyMaterial,
		Labels:    map[string]string{"key": strconv.Itoa(rInt)},
	}
	r.SetRName(name)
	return r
}
