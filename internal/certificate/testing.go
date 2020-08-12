package certificate

import (
	"context"
	"fmt"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/testsupport"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/testtemplate"
)

func init() {
	resource.AddTestSweepers(ResourceType, &resource.Sweeper{
		Name:         ResourceType,
		Dependencies: []string{},
		F:            Sweep,
	})
}

// Sweep removes all certificates from the Hetzner Cloud backend.
func Sweep(r string) error {
	client, err := testsupport.CreateClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	certificates, err := client.Certificate.All(ctx)
	if err != nil {
		return err
	}

	for _, cert := range certificates {
		if _, err := client.Certificate.Delete(ctx, cert); err != nil {
			return err
		}
	}

	return nil
}

// ByID returns a function that obtains a certificate by its ID.
func ByID(t *testing.T, cert *hcloud.Certificate) func(*hcloud.Client, int) bool {
	return func(c *hcloud.Client, id int) bool {
		found, _, err := c.Certificate.GetByID(context.Background(), id)
		if err != nil {
			t.Fatalf("find certificate %d: %v", id, err)
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

// DData defines the fields for the "testdata/d/hcloud_certificate"
// template.
type DData struct {
	testtemplate.DataCommon

	CertificateID   string
	CertificateName string
	LabelSelector   string
}

// TFID returns the data source identifier.
func (d *DData) TFID() string {
	return fmt.Sprintf("data.%s.%s", DataSourceType, d.RName())
}

// RData defines the fields for the "testdata/r/hcloud_certificate"
// template.
type RData struct {
	testtemplate.DataCommon

	Name        string
	PrivateKey  string
	Certificate string
	Labels      map[string]string
}

// TFID returns the resource identifier.
func (d *RData) TFID() string {
	return fmt.Sprintf("%s.%s", ResourceType, d.RName())
}

// NewRData creates data for a new certificate resource.
func NewRData(t *testing.T, name, org string) *RData {
	rCert, rKey, err := acctest.RandTLSCert(org)
	rInt := acctest.RandInt()
	if err != nil {
		t.Fatal(err)
	}
	r := &RData{
		Name:        name,
		PrivateKey:  rKey,
		Certificate: rCert,
		Labels:      map[string]string{"key": strconv.Itoa(rInt)},
	}
	r.SetRName(name)
	return r
}
