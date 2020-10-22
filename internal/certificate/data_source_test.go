package certificate_test

import (
	"fmt"
	"testing"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/certificate"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccHcloudDataSourceCertificateTest(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	res := certificate.NewRData(t, "datasource-test", "TFtestAcc")
	certificateByName := &certificate.DData{
		CertificateName: res.TFID() + ".name",
	}
	certificateByName.SetRName("certificate_by_name")
	certificateByID := &certificate.DData{
		CertificateID: res.TFID() + ".id",
	}
	certificateByID.SetRName("certificate_by_id")
	certificateBySel := &certificate.DData{
		LabelSelector: fmt.Sprintf("key=${%s.labels[\"key\"]}", res.TFID()),
	}
	certificateBySel.SetRName("certificate_by_sel")

	resource.Test(t, resource.TestCase{
		PreCheck:     testsupport.AccTestPreCheck(t),
		Providers:    testsupport.AccTestProviders(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(certificate.ResourceType, certificate.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_certificate", res,
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_certificate", res,
					"testdata/d/hcloud_certificate", certificateByName,
					"testdata/d/hcloud_certificate", certificateByID,
					"testdata/d/hcloud_certificate", certificateBySel,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(certificateByName.TFID(),
						"name", fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt)),
					resource.TestCheckResourceAttr(certificateByName.TFID(), "certificate", res.Certificate),

					resource.TestCheckResourceAttr(certificateByID.TFID(),
						"name", fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt)),
					resource.TestCheckResourceAttr(certificateByID.TFID(), "certificate", res.Certificate),

					resource.TestCheckResourceAttr(certificateBySel.TFID(),
						"name", fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt)),
					resource.TestCheckResourceAttr(certificateBySel.TFID(), "certificate", res.Certificate),
				),
			},
		},
	})
}
