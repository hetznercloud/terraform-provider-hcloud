package certificate_test

import (
	"fmt"
	"testing"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/certificate"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/e2etests"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestCertificateResource_Uploaded_Basic(t *testing.T) {
	var cert hcloud.Certificate

	res := certificate.NewUploadedRData(t, "basic-cert", "TFAccTests")
	resRenamed := &certificate.RDataUploaded{Name: res.Name + "-renamed", PrivateKey: res.PrivateKey, Certificate: res.Certificate}
	resRenamed.SetRName(res.Name)
	tmplMan := testtemplate.Manager{}
	// Not parallel because number of certificates per domain is limited
	resource.Test(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(certificate.UploadedResourceType, certificate.ByID(t, &cert)),
		Steps: []resource.TestStep{
			{
				// Create a new Certificate using the required values
				// only.
				Config: tmplMan.Render(t, "testdata/r/hcloud_uploaded_certificate", res),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), certificate.ByID(t, &cert)),
					resource.TestCheckResourceAttr(res.TFID(), "name",
						fmt.Sprintf("basic-cert--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res.TFID(), "private_key", res.PrivateKey),
					resource.TestCheckResourceAttr(res.TFID(), "certificate", res.Certificate),
				),
			},
			{
				// Update the Certificate created in the previous step by
				// setting all optional fields and renaming the Certificate.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_uploaded_certificate", resRenamed,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resRenamed.TFID(), "name",
						fmt.Sprintf("basic-cert-renamed--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(resRenamed.TFID(), "private_key", res.PrivateKey),
					resource.TestCheckResourceAttr(resRenamed.TFID(), "certificate", res.Certificate),
				),
			},
		},
	})
}

func TestCertificateResource_Uploaded_ChangeCertRequiresNewResource(t *testing.T) {
	var cert, newCert hcloud.Certificate

	res := certificate.NewUploadedRData(t, "basic-cert", "TFAccTests")

	rCert, rKey, err := testsupport.RandTLSCert("TFAccTests")
	if err != nil {
		t.Fatalf("%s", err)
	}
	resOtherCert := &certificate.RDataUploaded{Name: res.Name, PrivateKey: rKey, Certificate: rCert}
	resOtherCert.SetRName(res.Name)
	tmplMan := testtemplate.Manager{}
	// Not parallel because number of certificates per domain is limited
	resource.Test(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(certificate.UploadedResourceType, certificate.ByID(t, &cert)),
		Steps: []resource.TestStep{
			{
				// Create a new Certificate using the required values
				// only.
				Config: tmplMan.Render(t, "testdata/r/hcloud_uploaded_certificate", res),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), certificate.ByID(t, &cert)),
					resource.TestCheckResourceAttr(res.TFID(), "name",
						fmt.Sprintf("basic-cert--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res.TFID(), "private_key", res.PrivateKey),
					resource.TestCheckResourceAttr(res.TFID(), "certificate", res.Certificate),
				),
			},
			{
				// Update the Certificate created in the previous step by
				// setting all optional fields and renaming the Certificate.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_uploaded_certificate", resOtherCert,
				),
				Check: resource.ComposeAggregateTestCheckFunc(

					testsupport.CheckResourceExists(res.TFID(), certificate.ByID(t, &newCert)),
					resource.TestCheckResourceAttr(resOtherCert.TFID(), "name",
						fmt.Sprintf("basic-cert--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(resOtherCert.TFID(), "private_key", rKey),
					resource.TestCheckResourceAttr(resOtherCert.TFID(), "certificate", rCert),
					testsupport.LiftTCF(isAnotherCert(&newCert, &cert)),
				),
			},
		},
	})
}

func TestCertificateResource_Managed_Basic(t *testing.T) {
	var cert hcloud.Certificate

	res := certificate.NewManagedRData(t, "basic-managed-cert", []string{
		fmt.Sprintf("tftest-%d.hc-certs.de", acctest.RandInt()),
	})
	resRenamed := &certificate.RDataManaged{
		Name:        res.Name + "-renamed",
		DomainNames: res.DomainNames,
	}
	resRenamed.SetRName(res.Name)

	tmplMan := testtemplate.Manager{}
	// Not parallel because number of certificates per domain is limited
	resource.Test(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(certificate.ManagedResourceType, certificate.ByID(t, &cert)),
		Steps: []resource.TestStep{
			{
				// Create a new Certificate using the required values
				// only.
				Config: tmplMan.Render(t, "testdata/r/hcloud_managed_certificate", res),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), certificate.ByID(t, &cert)),
					resource.TestCheckResourceAttr(res.TFID(), "name",
						fmt.Sprintf("basic-managed-cert--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res.TFID(), "type", "managed"),
					resource.TestCheckResourceAttr(res.TFID(), "domain_names.0", res.DomainNames[0]),
				),
			},
			{
				// Update the Certificate created in the previous step by
				// setting all optional fields and renaming the Certificate.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_managed_certificate", resRenamed,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resRenamed.TFID(), "name",
						fmt.Sprintf("basic-managed-cert-renamed--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res.TFID(), "domain_names.0", res.DomainNames[0]),
				),
			},
		},
	})
}

func isAnotherCert(newCert *hcloud.Certificate, oldCert *hcloud.Certificate) func() error {
	return func() error {
		if newCert.ID == oldCert.ID {
			return fmt.Errorf("new cert is the same as old cert %d", oldCert.ID)
		}
		return nil
	}
}
