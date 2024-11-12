package certificate_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/certificate"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
)

func TestEqualCert_EqualCertificates(t *testing.T) {
	tests := []struct {
		name string
		cns  []string
	}{
		{
			name: "One certificate",
			cns:  []string{"aaa"},
		},
		{
			name: "Multiple certificates",
			cns:  []string{"aaa", "bbb"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			certs := make([]string, len(tt.cns))
			for i, cn := range tt.cns {
				cert, _, err := testsupport.RandTLSCert(cn)
				require.NoError(t, err)
				certs[i] = cert
			}

			cert := strings.Join(certs, "\n")
			res, err := certificate.EqualCert(cert, cert)
			require.NoError(t, err)

			assert.True(t, res, "Same certificates are expected to be equal")
		})
	}
}

func TestEqualCert_DifferentCertificates(t *testing.T) {
	tests := []struct {
		name   string
		cns    []string
		cnsAlt []string
	}{
		{
			name:   "one certificate - same cn",
			cns:    []string{"aaa"},
			cnsAlt: []string{"aaa"},
		},
		{
			name:   "multiple certificates - same cns",
			cns:    []string{"aaa", "bbb"},
			cnsAlt: []string{"aaa", "bbb"},
		},
		{
			name:   "first empty",
			cnsAlt: []string{"aaa", "bbb"},
		},
		{
			name: "second empty",
			cns:  []string{"aaa", "bbb"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			certs := make([]string, len(tt.cns))
			for i, cn := range tt.cns {
				cert, _, err := testsupport.RandTLSCert(cn)
				require.NoError(t, err)

				certs[i] = cert
			}

			certsAlt := make([]string, len(tt.cnsAlt))
			for i, cn := range tt.cnsAlt {
				cert, _, err := testsupport.RandTLSCert(cn)
				require.NoError(t, err)

				certsAlt[i] = cert
			}

			cert := strings.Join(certs, "\n")
			certAlt := strings.Join(certsAlt, "\n")
			res, err := certificate.EqualCert(cert, certAlt)
			require.NoError(t, err)

			assert.False(t, res, "Different certificates are expected to be unequal")
		})
	}
}
