package certificate_test

import (
	"fmt"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/certificate"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/testsupport"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/testtemplate"
)

func TestCertificateResource_Basic(t *testing.T) {
	var cert hcloud.Certificate

	res := certificate.NewRData(t, "basic-cert", "TFAccTests")
	resRenamed := &certificate.RData{Name: res.Name + "-renamed", PrivateKey: res.PrivateKey, Certificate: res.Certificate}
	resRenamed.SetRName(res.Name)
	tmplMan := testtemplate.Manager{}
	resource.Test(t, resource.TestCase{
		PreCheck:     testsupport.AccTestPreCheck(t),
		Providers:    testsupport.AccTestProviders(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(certificate.ResourceType, certificate.ByID(t, &cert)),
		Steps: []resource.TestStep{
			{
				// Create a new Certificate using the required values
				// only.
				Config: tmplMan.Render(t, "testdata/r/hcloud_certificate", res),
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
					"testdata/r/hcloud_certificate", resRenamed,
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
