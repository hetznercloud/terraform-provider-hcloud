package certificate_test

import (
	"fmt"
	"testing"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/certificate"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/e2etests"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccHcloudDataSourceCertificateTest(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	res := certificate.NewUploadedRData(t, "datasource-test", "TFtestAcc")
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(certificate.ResourceType, certificate.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_uploaded_certificate", res,
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_uploaded_certificate", res,
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

func TestAccHcloudDataSourceCertificateListTest(t *testing.T) {
	res := certificate.NewUploadedRData(t, "datasource-test", "TFtestAcc")

	certificateBySel := &certificate.DDataList{
		LabelSelector: fmt.Sprintf("key=${%s.labels[\"key\"]}", res.TFID()),
	}
	certificateBySel.SetRName("certificate_by_sel")

	allCertificateSel := &certificate.DDataList{}
	allCertificateSel.SetRName("all_certificates_sel")

	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(certificate.ResourceType, certificate.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_uploaded_certificate", res,
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_uploaded_certificate", res,
					"testdata/d/hcloud_certificates", certificateBySel,
					"testdata/d/hcloud_certificates", allCertificateSel,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(certificateBySel.TFID(), "certificates.*",
						map[string]string{
							"name":        fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt),
							"certificate": res.Certificate,
						},
					),

					resource.TestCheckTypeSetElemNestedAttrs(allCertificateSel.TFID(), "certificates.*",
						map[string]string{
							"name":        fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt),
							"certificate": res.Certificate,
						},
					),
				),
			},
		},
	})
}
