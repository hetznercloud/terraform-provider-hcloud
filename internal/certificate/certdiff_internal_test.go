package certificate

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
)

// TestParseCertificates_CertificateChain tries to parse a chain of PEM
// encoded certificates with interspersed and terminating data.
// See: https://github.com/hetznercloud/terraform-provider-hcloud/issues/359
func TestParseCertificates_CertificateChain(t *testing.T) {
	pem1, _, err := testsupport.RandTLSCert("example.com")
	assert.NoError(t, err)
	cert1, err := parseCertificates(pem1)
	assert.NoError(t, err)

	pem2, _, err := testsupport.RandTLSCert("ca.example.com")
	if !assert.NoError(t, err) {
		return
	}
	cert2, err := parseCertificates(pem2)
	if !assert.NoError(t, err) {
		return
	}

	// Not really a certificate chain, but enough for our testing purposes
	chain := strings.Join([]string{pem1, pem2}, "\nIntermediate data\n") + "Terminating data"
	actual, err := parseCertificates(chain)
	assert.NoError(t, err)

	if !assert.Len(t, actual, 2) {
		return
	}
	assert.Equal(t, cert1[0], actual[0])
	assert.Equal(t, cert2[0], actual[1])
}
