package certificate

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func init() {
	resource.AddTestSweepers(UploadedResourceType, &resource.Sweeper{
		Name:         UploadedResourceType,
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

// DData defines the fields for the "testdata/d/hcloud_uploaded_certificate"
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

// RDataUploaded defines the fields for the "testdata/r/hcloud_uploaded_certificate"
// template.
type RDataUploaded struct {
	testtemplate.DataCommon

	Name        string
	PrivateKey  string
	Certificate string
	Labels      map[string]string
}

// NewUploadedRData creates data for a new certificate resource.
func NewUploadedRData(t *testing.T, name, org string) *RDataUploaded {
	rInt := acctest.RandInt()
	rCert, rKey, err := testsupport.RandTLSCert(org)
	if err != nil {
		t.Fatal(err)
	}
	r := &RDataUploaded{
		Name:        name,
		PrivateKey:  rKey,
		Certificate: rCert,
		Labels:      map[string]string{"key": strconv.Itoa(rInt)},
	}
	r.SetRName(name)
	return r
}

// TFID returns the resource identifier.
func (d *RDataUploaded) TFID() string {
	return fmt.Sprintf("%s.%s", UploadedResourceType, d.RName())
}

// RDataManaged defines the fields for the "testdata/r/hcloud_managed_certificate"
// template.
type RDataManaged struct {
	testtemplate.DataCommon

	Name        string
	DomainNames []string
	Labels      map[string]string
}

// NewManagedRData creates data for a new certificate resource.
func NewManagedRData(t *testing.T, name string, domainNames []string) *RDataManaged {
	r := &RDataManaged{
		Name:        name,
		DomainNames: domainNames,

		Labels: map[string]string{
			// Required for internal testing purposes.
			// DO NOT USE THIS LABEL OUTSIDE OF HETZNER CLOUD TESTS!
			//
			// Support for it may vanish any time we see fit.
			"HC-Use-Staging-CA": "true",
		},
	}
	r.SetRName(name)
	return r
}

// TFID returns the resource identifier.
func (d *RDataManaged) TFID() string {
	return fmt.Sprintf("%s.%s", ManagedResourceType, d.RName())
}
