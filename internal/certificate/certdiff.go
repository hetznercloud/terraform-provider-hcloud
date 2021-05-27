package certificate

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// EqualCert compares two PEM encoded X509 certificates.
//
// EqualCert returns an error if it was unable to parse any of the passed
// certificates.
func EqualCert(cert1, cert2 string) (bool, error) {
	const op = "certificate/EqualCert"

	if cert1 == "" || cert2 == "" {
		return false, nil
	}

	parsed1, err := parseCertificates(cert1)
	if err != nil {
		return false, fmt.Errorf("%s: cert1: %v", op, err)
	}
	parsed2, err := parseCertificates(cert2)
	if err != nil {
		return false, fmt.Errorf("%s: cert2: %v", op, err)
	}
	if len(parsed1) != len(parsed2) {
		return false, nil
	}

	hashes1 := calcCertHashes(parsed1)
	hashes2 := calcCertHashes(parsed2)
	for h := range hashes1 {
		if !hashes2[h] {
			return false, nil
		}
		delete(hashes1, h)
		delete(hashes2, h)
	}

	return len(hashes1) == 0 && len(hashes2) == 0, nil
}

func calcCertHashes(cs []*x509.Certificate) map[[sha256.Size]byte]bool {
	hs := make(map[[sha256.Size]byte]bool, len(cs))
	for _, c := range cs {
		h := sha256.Sum256(c.Raw)
		hs[h] = true
	}
	return hs
}

func parseCertificates(cert string) ([]*x509.Certificate, error) {
	const op = "certificate/parseCertificate"
	var certs []*x509.Certificate

	rest := []byte(cert)
	for len(rest) > 0 {
		var blk *pem.Block

		blk, rest = pem.Decode(rest)
		if blk == nil {
			// No (further) PEM data found. We are done.
			break
		}
		c, err := x509.ParseCertificate(blk.Bytes)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", op, err)
		}

		certs = append(certs, c)
	}

	return certs, nil
}
