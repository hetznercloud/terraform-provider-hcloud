package testsupport

import (
	"bytes"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"math/big"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/hetznercloud/hcloud-go/hcloud"
	tfhcloud "github.com/hetznercloud/terraform-provider-hcloud/hcloud"
)

const (
	// TestImage is the system image that is used in all tests
	TestImage = "ubuntu-20.04"

	// TestServerType is the default server type used in all tests
	TestServerType = "cx11"

	// TestLoadBalancerType is the default Load Balancer type used in all tests
	TestLoadBalancerType = "lb11"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

// AccTestProviders returns all providers used during acceptance testing.
func AccTestProviders() map[string]*schema.Provider {
	return map[string]*schema.Provider{
		"hcloud": tfhcloud.Provider(),
	}
}

// CreateClient creates a new *hcloud.Client which authenticates using the
// HCLOUD_TOKEN variable.
func CreateClient() (*hcloud.Client, error) {
	if os.Getenv("HCLOUD_TOKEN") == "" {
		return nil, fmt.Errorf("empty HCLOUD_TOKEN")
	}
	opts := []hcloud.ClientOption{
		hcloud.WithToken(os.Getenv("HCLOUD_TOKEN")),
	}
	return hcloud.NewClient(opts...), nil
}

// AccTestPreCheck checks if all conditions for an acceptance test are
// met.
func AccTestPreCheck(t *testing.T) func() {
	return func() {
		if v := os.Getenv("HCLOUD_TOKEN"); v == "" {
			t.Fatal("HCLOUD_TOKEN must be set for acceptance tests")
		}
	}
}

// RandTLSCert generates a self-signed TLS certificate with a newly created
// private key, and returns both the cert and the private key PEM encoded.
func RandTLSCert(orgName string) (string, string, error) {
	template := &x509.Certificate{
		SerialNumber: big.NewInt(int64(acctest.RandInt())),
		Subject: pkix.Name{
			Organization: []string{orgName},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	privateKey, privateKeyPEM, err := genPrivateKey()
	if err != nil {
		return "", "", err
	}

	cert, err := x509.CreateCertificate(crand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return "", "", err
	}

	certPEM, err := pemEncode(cert, "CERTIFICATE")
	if err != nil {
		return "", "", err
	}

	return certPEM, privateKeyPEM, nil
}

func genPrivateKey() (*rsa.PrivateKey, string, error) {
	privateKey, err := rsa.GenerateKey(crand.Reader, 2048)
	if err != nil {
		return nil, "", err
	}

	privateKeyPEM, err := pemEncode(x509.MarshalPKCS1PrivateKey(privateKey), "RSA PRIVATE KEY")
	if err != nil {
		return nil, "", err
	}

	return privateKey, privateKeyPEM, nil
}

func pemEncode(b []byte, block string) (string, error) {
	var buf bytes.Buffer
	pb := &pem.Block{Type: block, Bytes: b}
	if err := pem.Encode(&buf, pb); err != nil {
		return "", err
	}

	return buf.String(), nil
}
