package bootstrap

import (
	"crypto/x509"
	"errors"
)

// VerifyCertificateChain verifies server certificates without checking a hostname.
func VerifyCertificateChain(certificates []*x509.Certificate, roots *x509.CertPool) error {
	if len(certificates) == 0 {
		return errors.New("database server did not provide a certificate")
	}
	intermediates := x509.NewCertPool()
	for _, cert := range certificates[1:] {
		intermediates.AddCert(cert)
	}
	_, err := certificates[0].Verify(x509.VerifyOptions{
		Roots: roots, Intermediates: intermediates,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	})
	return err
}
