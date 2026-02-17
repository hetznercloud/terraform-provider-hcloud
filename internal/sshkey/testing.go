package sshkey

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

// GetAPIResource returns a [testsupport.GetAPIResourceFunc] for [hcloud.SSHKey].
func GetAPIResource() testsupport.GetAPIResourceFunc[hcloud.SSHKey] {
	return func(c *hcloud.Client, attrs map[string]string) (*hcloud.SSHKey, error) {
		result, _, err := c.SSHKey.Get(context.Background(), attrs["id"])
		return result, err
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

// DDataList defines the fields for the "testdata/d/hcloud_ssh_keys"
// template.
type DDataList struct {
	testtemplate.DataCommon
	LabelSelector string
}

// TFID returns the data source identifier.
func (d *DDataList) TFID() string {
	return fmt.Sprintf("data.%s.%s", DataSourceListType, d.RName())
}

// RData defines the fields for the "testdata/r/hcloud_ssh_key"
// template.
type RData struct {
	testtemplate.DataCommon

	Name       string
	PublicKey  string
	PrivateKey string // nolint: gosec
	Labels     map[string]string
}

// TFID returns the resource identifier.
func (d *RData) TFID() string {
	return fmt.Sprintf("%s.%s", ResourceType, d.RName())
}

// NewRData creates data for a new sshkey resource.
func NewRData(t *testing.T, name string) *RData {
	publicKeyMaterial, privateKeyMaterial, err := acctest.RandSSHKeyPair("hcloud@ssh-acceptance-test")
	rInt := acctest.RandInt()
	if err != nil {
		t.Fatal(err)
	}
	r := &RData{
		Name:       name,
		PublicKey:  publicKeyMaterial,
		PrivateKey: privateKeyMaterial,
		Labels:     map[string]string{"key": strconv.Itoa(rInt)},
	}
	r.SetRName(name)
	return r
}
