package hcloud

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

// Never use them in a production environment.
const (
	testDataSourceHcloudCertificate_PrivateKey = `
-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEApfVaqZ1R2A7Ub9HX9G9+sUKlUeJpEq21XxtxKeq4b2PykSnA
V2ylLGFRQkg+dlbcEtOlW7tVZLSmX1/5MosnHv2WrYtgkrIiEVe0ytyt+FDwjEnS
EZNwkPLwvVcdHJCQ04t9QS4pWb6CcwmAelA+mP03QMdsI1Vh8xCDjLIMYR0sfM1e
6UC7WCCuZqcMsF3LXBG7azFHlJ8JuFD6Zwpq2uoh++SKPgQdXNRf+ChXszgmnIEm
qzx1mmBYFC2J7F8P7To29qVQmv74kmJQVukKUM6MauGVavlOaGaBJLnLJzK+eFRb
GXIaXZPm4nFxQB8Glsv5Cgm7sY+Rt70kbmIdJwIDAQABAoIBADvR988hxOmTVuHD
iwsx0UIF2t9sNCGmebUBiEXqtHZ6WWoX2ZvprpQTtB2nOtSfNI6YuYcaPIqjT0Eb
sBCW9hAGqnR9w9681OjZa3KgHxld24dF6LGrKq0k1L/7EaRFO9707b477G3L7OuY
ZjYwjI/K3kq8V29ZDIh30GM7npi2Pa4VaM+eSrILwHaN/RHLC6cq/XY+6ROo980e
0EPwTXBdCPOmEa5fFMsqnUkdfpfsnqZc92TzrI16OuKDzjDnHyvv/4uQ9jmg2wya
krzgIkEexJkcGP+ykq3ZZwNgRecojglKb/nafA+qlesGJfs0iBRhD8KRhEATYW54
Yv9wwcECgYEA1X4u0Hi/79ZYvF19abXSEg1F574t8XCWvEHCfbEGrWGcLChitAdl
Ee+21+IGh1d5zdJ/hRppSWyaDklvB7XeHeZOzvkgRe1RKGgMBwad+M85gRZyODTD
VtU+ocIFchF+IqTXXS15hFPBk0lSQq7dZ/c10qZ1mH9bQIyU5PPrlWECgYEAxwBR
3WWUEgTb+jQX8a24x2Hy6KG/3Ms5PK0BXWquYngeQLkx6Tz5aWavRajB9vc32s9z
XfHCm7tv09tWvNrBUNag5ba/z/2ArlPePgybfNI66rZLpP9heSAWtVzLlxExQKgw
kTPvEnzWX4tsWfRaq+OCuo4WvCF8riwlNWQet4cCgYEAudorLtivXj6e6PwKHWhn
A8gCPwfUPwbgcepdQcZGJdF/fwF5S3fUiJTB+5WMUW3ZX1AMKvcfCQg95IoQ2gl8
31KK8Kr3aWh66k4JimQ8SUk8qh+8NynXk1P4PiEFVJPd1pLh2P+pdYTkUy/VKK/J
lqQiesrmPGdCLSM0y0t8noECgYBqKHiDg9mulxsGaV3Qllz5N/5OLWNdlKfu/1e4
Dt4CN5Pj8Sd4BggDOz0LCxCV/6GzP3GKzxqC20W3nc2yp3vy9NwWTxwaB2DrHmBz
d2RG/Rti9GZ8GaRU6lJS47LT3t8IX/CwtSS3FxOBGq5telYYViD6BiyIpdCOVYxv
4/4i5wKBgCXA8mKhzMngHCrMhruD76n2/EJ5NWLVDuFsKRosg+gGof1ZbY9IgEiL
tjxJabLYWc2cVzeOUM8/aWfoWxJnwlL8whtNovlxZqhsdznen5rBu8mx6iGCRptO
DkkILLSeDIyL7wkmt4jqqAqJNoiYnq8mqXjKrf+E/opWxLpp+4Rj
-----END RSA PRIVATE KEY-----
`
	testDataSourceHcloudCertificate_Certificate = `
-----BEGIN CERTIFICATE-----
MIIDMDCCAhigAwIBAgIICiysUjsqSqgwDQYJKoZIhvcNAQELBQAwIDEeMBwGA1UE
AxMVbWluaWNhIHJvb3QgY2EgMWJkZGYyMB4XDTIwMDQwMzA3MjkwNVoXDTIyMDUw
MzA3MjkwNVowFjEUMBIGA1UEAxMLZXhhbXBsZS5jb20wggEiMA0GCSqGSIb3DQEB
AQUAA4IBDwAwggEKAoIBAQCl9VqpnVHYDtRv0df0b36xQqVR4mkSrbVfG3Ep6rhv
Y/KRKcBXbKUsYVFCSD52VtwS06Vbu1VktKZfX/kyiyce/Zati2CSsiIRV7TK3K34
UPCMSdIRk3CQ8vC9Vx0ckJDTi31BLilZvoJzCYB6UD6Y/TdAx2wjVWHzEIOMsgxh
HSx8zV7pQLtYIK5mpwywXctcEbtrMUeUnwm4UPpnCmra6iH75Io+BB1c1F/4KFez
OCacgSarPHWaYFgULYnsXw/tOjb2pVCa/viSYlBW6QpQzoxq4ZVq+U5oZoEkucsn
Mr54VFsZchpdk+bicXFAHwaWy/kKCbuxj5G3vSRuYh0nAgMBAAGjeDB2MA4GA1Ud
DwEB/wQEAwIFoDAdBgNVHSUEFjAUBggrBgEFBQcDAQYIKwYBBQUHAwIwDAYDVR0T
AQH/BAIwADAfBgNVHSMEGDAWgBThDNtx3hjju6iDko9R44XZwKQLMDAWBgNVHREE
DzANggtleGFtcGxlLmNvbTANBgkqhkiG9w0BAQsFAAOCAQEATfl0T/tkQLyGxTea
w5cwOry5/i0UG+8kdj3VfC7WFEVQzjAmL9BdOWf1HEdq2BAtqBioLovE8P/Ob9CB
cAaArtwvsRdVjwjPhLWrdhwI4F6Ihxxkk6cY5FRb1wxRIv2BL+iUeabS2Xzbv6gv
9XTQ6Kpt8fNdxjeaJmeUKy5gzdVYsqYbZwDp9fsVvBEuKjJTr1Z5rkVEx9ldjGyk
v8JhAZIcpfq0GiwzP3EdgQvgblTHR+TmS3sSk6VMYF/ikEGn41fp0RS7JvdccHoD
/uSi8kYGCt24glyCE9artV8huih1/xw5MuWjxy7iu/wKA3aGzESy2KYJjDtYXXrZ
3rO/TA==
-----END CERTIFICATE-----
`
	testDataSourceHcloudCertificate_Fingerprint    = "D0:D3:99:D2:E2:31:75:24:D0:1E:79:83:BB:DD:EC:6B:6B:C0:99:30:4E:09:F8:B2:42:C8:47:28:22:D5:D4:B2"
	testDataSourceHcloudCertificate_NotValidBefore = "2020-04-03T07:29:05Z"
	testDataSourceHcloudCertificate_NotValidAfter  = "2022-05-03T07:29:05Z"
)

func init() {
	resource.AddTestSweepers("data_source_certificate", &resource.Sweeper{
		Name: "hcloud_certificate_data_source",
		F:    testSweepCertificates,
	})
}

func TestAccHcloudDataSourceCertificate(t *testing.T) {
	var certificate hcloud.Certificate
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccHcloudPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHcloudCertificateDataSourceConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					// Basic Resource
					testAccHcloudCertificateExists("hcloud_certificate.test-certificate", &certificate),
					resource.TestCheckResourceAttr(
						"hcloud_certificate.test-certificate", "name", fmt.Sprintf("test-certificate-%d", rInt)),
					resource.TestCheckResourceAttr("hcloud_certificate.test-certificate", "certificate", strings.TrimPrefix(testDataSourceHcloudCertificate_Certificate, "\n")),
					resource.TestCheckResourceAttr("hcloud_certificate.test-certificate", "domain_names.#", "1"),
					resource.TestCheckResourceAttr("hcloud_certificate.test-certificate", "domain_names.0", "example.com"),
					resource.TestCheckResourceAttr("hcloud_certificate.test-certificate", "fingerprint", testDataSourceHcloudCertificate_Fingerprint),
					resource.TestCheckResourceAttr("hcloud_certificate.test-certificate", "not_valid_before", testDataSourceHcloudCertificate_NotValidBefore),
					resource.TestCheckResourceAttr("hcloud_certificate.test-certificate", "not_valid_after", testDataSourceHcloudCertificate_NotValidAfter),
					// Certificate via Name
					resource.TestCheckResourceAttr(
						"data.hcloud_certificate.certificate_name", "name", fmt.Sprintf("test-certificate-%d", rInt)),
					resource.TestCheckResourceAttr("data.hcloud_certificate.certificate_name", "certificate", strings.TrimPrefix(testDataSourceHcloudCertificate_Certificate, "\n")),
					resource.TestCheckResourceAttr("data.hcloud_certificate.certificate_name", "domain_names.#", "1"),
					resource.TestCheckResourceAttr("data.hcloud_certificate.certificate_name", "domain_names.0", "example.com"),
					resource.TestCheckResourceAttr("data.hcloud_certificate.certificate_name", "fingerprint", testDataSourceHcloudCertificate_Fingerprint),
					resource.TestCheckResourceAttr("data.hcloud_certificate.certificate_name", "not_valid_before", testDataSourceHcloudCertificate_NotValidBefore),
					resource.TestCheckResourceAttr("data.hcloud_certificate.certificate_name", "not_valid_after", testDataSourceHcloudCertificate_NotValidAfter),
					// Certificate via ID
					resource.TestCheckResourceAttr(
						"data.hcloud_certificate.certificate_id", "name", fmt.Sprintf("test-certificate-%d", rInt)),
					resource.TestCheckResourceAttr("data.hcloud_certificate.certificate_id", "certificate", strings.TrimPrefix(testDataSourceHcloudCertificate_Certificate, "\n")),
					resource.TestCheckResourceAttr("data.hcloud_certificate.certificate_id", "domain_names.#", "1"),
					resource.TestCheckResourceAttr("data.hcloud_certificate.certificate_id", "domain_names.0", "example.com"),
					resource.TestCheckResourceAttr("data.hcloud_certificate.certificate_id", "fingerprint", testDataSourceHcloudCertificate_Fingerprint),
					resource.TestCheckResourceAttr("data.hcloud_certificate.certificate_id", "not_valid_before", testDataSourceHcloudCertificate_NotValidBefore),
					resource.TestCheckResourceAttr("data.hcloud_certificate.certificate_id", "not_valid_after", testDataSourceHcloudCertificate_NotValidAfter),
					// Certificate via Label Selector
					resource.TestCheckResourceAttr(
						"data.hcloud_certificate.certificate_label", "name", fmt.Sprintf("test-certificate-%d", rInt)),
					resource.TestCheckResourceAttr("data.hcloud_certificate.certificate_label", "certificate", strings.TrimPrefix(testDataSourceHcloudCertificate_Certificate, "\n")),
					resource.TestCheckResourceAttr("data.hcloud_certificate.certificate_label", "domain_names.#", "1"),
					resource.TestCheckResourceAttr("data.hcloud_certificate.certificate_label", "domain_names.0", "example.com"),
					resource.TestCheckResourceAttr("data.hcloud_certificate.certificate_label", "fingerprint", testDataSourceHcloudCertificate_Fingerprint),
					resource.TestCheckResourceAttr("data.hcloud_certificate.certificate_label", "not_valid_before", testDataSourceHcloudCertificate_NotValidBefore),
					resource.TestCheckResourceAttr("data.hcloud_certificate.certificate_label", "not_valid_after", testDataSourceHcloudCertificate_NotValidAfter),
				),
			},
		},
	})
}

func testAccCheckHcloudCertificateDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
variable "labels" {
  type = "map"
  default = {
    "key" = "%d"
  }
}
resource "hcloud_certificate" "test-certificate" {
    name = "test-certificate-%d"
    private_key =<<EOT%sEOT
    certificate =<<EOT%sEOT
    labels     = "${var.labels}"
}

data "hcloud_certificate" "certificate_name" {
  name = "${hcloud_certificate.test-certificate.name}"
}
data "hcloud_certificate" "certificate_id" {
  id =  "${hcloud_certificate.test-certificate.id}"
}
data "hcloud_certificate" "certificate_label" {
  with_selector =  "key=${hcloud_certificate.test-certificate.labels["key"]}"
}
`, rInt, rInt, testDataSourceHcloudCertificate_PrivateKey, testDataSourceHcloudCertificate_Certificate)
}
