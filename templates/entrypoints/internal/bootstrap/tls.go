package bootstrap

import (
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// LoadCAPool extends the system certificate pool with PEM certificates from a
// directory. OpenSSL hash links are accepted; duplicate certificates are safe.
func LoadCAPool(directory string) (*x509.CertPool, error) {
	pool, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("load system CA certificates: %w", err)
	}
	if directory == "" {
		return pool, nil
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("read CA directory %q: %w", directory, err)
	}
	for _, entry := range entries {
		path := filepath.Join(directory, entry.Name())
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("stat CA entry %q: %w", path, err)
		}
		if !info.Mode().IsRegular() {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read CA certificate %q: %w", path, err)
		}
		if !pool.AppendCertsFromPEM(data) {
			LogWarn("** No PEM certificates parsed from CA file %q", path)
		}
	}
	return pool, nil
}

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
