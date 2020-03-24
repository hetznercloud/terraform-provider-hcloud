package hcloud

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

// Never use them in a production environment.
const (
	testResourceHcloudCertificate_PrivateKey = `
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEArTKYunlOu+f+9WjGtXMw8zpDRM1ZBgIVhWvTsnBzN6OUwmzw
31MWGRlGe90aIsFoNsHAYd2EkaSx2tLMSbuQrnJCrhfrrpqg9t306/pcedZlCMlm
XPaB130dn+6MCTQIIRi4W6/IPwU2exeVCN1fGOHnD0v1eHkZE0dFqNNCOx2gLdQ2
0DP/NT11PYBGv9reoOddQlpnZHFikViH6cd0Y1bwtF4M1qi944KTHpCsjUHue0m9
9GONtP73sN5hGGuFuLQ08Mrh7cF8uJiWgaKpShuFtGE9XjCuDt3mGIN1VqBVofof
qEgfp4c3wgbO4xJU848zA688YueX3npqflwEZwIDAQABAoIBADDBio8L8Scikvpg
yXdw+vmtkBxBNmtjkM1jYk/cKgMisJDe/BvSFulv3RCnWGEqUvz/I/oo0tXxKAQy
zUGAZKVHExBROY6IhwGX0AfSDdBak0ya7Y8D8d9IoFtSWueIVzWB8PwWiud6vzB9
nf3F26x0g4gh2PNWG8H6kViKSB29rbi6Uss7oz8eP5A/8JIMjz1GGZmHr9huthE+
aFTKXoUXoXF+I8LkD0sinhxITjD6vMYhursXvf3b0bP25Wg4fWG3JSX+pUiCBrTM
zV+5XHh6QRqx6t/gsQfMQlUzLsoYTiYeGQ1lhu9a6CVdkAeIsX9n9Gp6QxEXSHN/
sYhDmiECgYEA0WlccG3pGclYEgQ20AYtCxdyOAqnlCF98Jut/PyK2zHY5YZsuM1x
9eeMJUiwdI11iT7x8mlGk3nWxN+HfnkzpLel3P3a3qgoygDkXrAPqPYcc41HfXNo
vahTts5qhy3cRcPYIEi1YwIkTPO7AdOSKgGQgXmjuqCFKxmAoBFT7jECgYEA07q9
d6Ehosq35/u5w/W0hkNoCMNJvIRXXCWkoypdc+NThRfUefHusT0Vl2HBRLvHYC/H
Mbq6uD3vU5UKGop9I9eoBezrDh7u8uy3q7fkrFosz/NbORmQNsX2Ljbw9BSaQ4lR
lxZwKr2Wp3mrfCzr/ArwFZ8ZdWsnjhM4lGEN/hcCgYEApgxDbirY0Meke/S3ec/L
26WlveZE5uJ/uE/ZcGbXu+MUtzsV7puJJ5GIwO+Ya3LXphIxSyRLABl2QPl1uMVm
O9AbRtZLvI3eef6nFqXIZRNxj/aQn1rpzKkyaBvYwIOOzAr0zvSYT3+dRR9mQ5Z9
qa0/5kqLlyo9LeW05jeXM6ECgYAdocDqgS6H7f8XBG/XMQf20nA46bvkGlFvoAUO
oNs7YNFLiy49ctKJE5d1/ERkLjOVDpq+JvgC2Qgplm43kLI61e+6BJJRA5tFfEOo
ULA8PtKOt+xIbX91avctOJs4TbnZQdqdXpKMKMRw4+JQGqlcONuo6v9RI5IBnEcK
3RpsOQKBgQCenVLepqyVd63GxLeJTawPF7Dm31yKFZXnj5yV2ORmBmucX3nunj6h
kAOHazSFWzz+4Rm2IQUTgUHJB17CWeuAjbYmHkYmHVsRQxlC0elfzM2APeXxWlfG
cK0Y1gPNkLU5ax2shOekyv2x/uOQyJmIPVBfLr570+q5nOLojEwbwQ==
-----END RSA PRIVATE KEY-----
`
	testResourceHcloudCertificate_Certificate = `
-----BEGIN CERTIFICATE-----
MIIDMDCCAhigAwIBAgIINoLWZ/ADzeswDQYJKoZIhvcNAQELBQAwIDEeMBwGA1UE
AxMVbWluaWNhIHJvb3QgY2EgMWJkZGYyMB4XDTIwMDQwMjEyMDExOFoXDTIyMDUw
MjEyMDExOFowFjEUMBIGA1UEAxMLZXhhbXBsZS5vcmcwggEiMA0GCSqGSIb3DQEB
AQUAA4IBDwAwggEKAoIBAQCtMpi6eU675/71aMa1czDzOkNEzVkGAhWFa9OycHM3
o5TCbPDfUxYZGUZ73RoiwWg2wcBh3YSRpLHa0sxJu5CuckKuF+uumqD23fTr+lx5
1mUIyWZc9oHXfR2f7owJNAghGLhbr8g/BTZ7F5UI3V8Y4ecPS/V4eRkTR0Wo00I7
HaAt1DbQM/81PXU9gEa/2t6g511CWmdkcWKRWIfpx3RjVvC0XgzWqL3jgpMekKyN
Qe57Sb30Y420/vew3mEYa4W4tDTwyuHtwXy4mJaBoqlKG4W0YT1eMK4O3eYYg3VW
oFWh+h+oSB+nhzfCBs7jElTzjzMDrzxi55feemp+XARnAgMBAAGjeDB2MA4GA1Ud
DwEB/wQEAwIFoDAdBgNVHSUEFjAUBggrBgEFBQcDAQYIKwYBBQUHAwIwDAYDVR0T
AQH/BAIwADAfBgNVHSMEGDAWgBThDNtx3hjju6iDko9R44XZwKQLMDAWBgNVHREE
DzANggtleGFtcGxlLm9yZzANBgkqhkiG9w0BAQsFAAOCAQEAeNl9iUSXUkvTzkHc
Q1I9IO2IFKeLsBGMu2DVMkZd28yY+LsBm3FZDdLgfY6naCSXVmaPfCpjGIWp881h
UMgJFCwWRRE0nNlyVy+FG1thPGaaL17Pjpxlv0FLchhRMVZ3wrdnM6PgRdS36oiE
69NURrix4A9BSopxilrIlxEKGy5eOHLflgRk5Ie0P70Cs3PG4l+VtXiCP+uAJY9r
pPOqvKM9yzYgTlLEcROKzu1bl6oKS+fXH+5QBmBVExCOFkkF+G6rsNPS+W0vU5ls
tJBOrw3T2DSC31JnMaTM0UKh+yj1zVcQU57PqaB9OnY0FXAUzKGI/ojMFAkLW9e2
RN5rsQ==
-----END CERTIFICATE-----
`
	testResourceHcloudCertificate_Fingerprint    = "05:36:3E:88:65:45:8A:36:BC:4B:FD:95:34:B9:27:A3:7C:E1:AC:92:E6:82:C3:FC:41:F2:64:9C:E8:98:28:04"
	testResourceHcloudCertificate_NotValidBefore = "2020-04-02T12:01:18Z"
	testResourceHcloudCertificate_NotValidAfter  = "2022-05-02T12:01:18Z"
)

func init() {
	resource.AddTestSweepers("resource_hcloud_certificate", &resource.Sweeper{
		Name: "hcloud_certificate",
		F:    testSweepCertificates,
	})
}

func TestAccHcloudCertificate(t *testing.T) {
	var certificate hcloud.Certificate
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccHcloudPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCertificateConfig_Basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCertificateExists("hcloud_certificate.test-certificate", &certificate),
					resource.TestCheckResourceAttr(
						"hcloud_certificate.test-certificate", "name", fmt.Sprintf("test-certificate-%d", rInt)),
					resource.TestCheckResourceAttr("hcloud_certificate.test-certificate", "certificate", strings.TrimPrefix(testResourceHcloudCertificate_Certificate, "\n")),
					resource.TestCheckResourceAttr("hcloud_certificate.test-certificate", "domain_names.#", "1"),
					resource.TestCheckResourceAttr("hcloud_certificate.test-certificate", "domain_names.0", "example.org"),
					resource.TestCheckResourceAttr("hcloud_certificate.test-certificate", "fingerprint", testResourceHcloudCertificate_Fingerprint),
					resource.TestCheckResourceAttr("hcloud_certificate.test-certificate", "not_valid_before", testResourceHcloudCertificate_NotValidBefore),
					resource.TestCheckResourceAttr("hcloud_certificate.test-certificate", "not_valid_after", testResourceHcloudCertificate_NotValidAfter),
				),
			},
			{
				Config: testAccHcloudCertificateConfig_WithLabels(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCertificateExists("hcloud_certificate.test-certificate", &certificate),
					resource.TestCheckResourceAttr("hcloud_certificate.test-certificate", "labels.%", "2"),
					resource.TestCheckResourceAttr("hcloud_certificate.test-certificate", "labels.key_1", "value_1"),
					resource.TestCheckResourceAttr("hcloud_certificate.test-certificate", "labels.key_2", "value_2"),
				),
			},
			{
				Config: testAccHcloudCertificateConfig_Renamed(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCertificateExists("hcloud_certificate.test-certificate", &certificate),
					resource.TestCheckResourceAttr(
						"hcloud_certificate.test-certificate", "name", fmt.Sprintf("test-certificate-renamed-%d", rInt)),
				),
			},
		},
	})
}

func testCertificateDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*hcloud.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "hcloud_load_certificate" {
			continue
		}

		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Certificate id is not an int: %v", err)
		}

		certificate, _, err := client.Certificate.GetByID(context.Background(), id)
		if err != nil {
			return fmt.Errorf(
				"Error checking if Certificate (%s) is deleted: %v",
				rs.Primary.ID, err)
		}
		if certificate != nil {
			return fmt.Errorf("Load Balancer (%s) has not been deleted", rs.Primary.ID)
		}
	}
	return nil
}

func testAccHcloudCertificateConfig_Basic(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_certificate" "test-certificate" {
	name = "test-certificate-%d"
	private_key =<<EOT%sEOT
	certificate =<<EOT%sEOT
}
`, rInt, testResourceHcloudCertificate_PrivateKey, testResourceHcloudCertificate_Certificate)
}

func testAccHcloudCertificateConfig_WithLabels(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_certificate" "test-certificate" {
	name = "test-certificate-%d"
	private_key =<<EOT%sEOT
	certificate =<<EOT%sEOT
	labels = {
		key_1 = "value_1"
		key_2 = "value_2"
	}
}
`, rInt, testResourceHcloudCertificate_PrivateKey, testResourceHcloudCertificate_Certificate)
}

func testAccHcloudCertificateConfig_Renamed(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_certificate" "test-certificate" {
	name = "test-certificate-renamed-%d"
	private_key =<<EOT%sEOT
	certificate =<<EOT%sEOT
}
`, rInt, testResourceHcloudCertificate_PrivateKey, testResourceHcloudCertificate_Certificate)
}

func testAccHcloudCertificateExists(n string, certificate *hcloud.Certificate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("resource id not set")
		}

		client := testAccProvider.Meta().(*hcloud.Client)
		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return err
		}
		cert, _, err := client.Certificate.GetByID(context.Background(), id)
		if err != nil {
			return err
		}
		if cert == nil {
			return fmt.Errorf("certificate not found: %d", id)
		}
		*certificate = *cert
		return nil
	}
}

func testSweepCertificates(_ string) error {
	client, err := createClient()
	if err != nil {
		return err
	}
	certs, err := client.Certificate.All(context.Background())
	if err != nil {
		return err
	}
	for _, cert := range certs {
		if _, err := client.Certificate.Delete(context.Background(), cert); err != nil {
			return err
		}
	}
	return nil
}
