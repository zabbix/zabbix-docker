package mysql

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"
)

func (db *Database) tlsConfig() (*tls.Config, error) {
	if db.env["ZBX_DB_ENCRYPTION"] == "true" {
		db.env["ZBX_DBTLSCONNECT"] = "required"
	}

	mode := strings.ToLower(db.env["ZBX_DBTLSCONNECT"])
	if mode == "" {
		return nil, nil
	}
	if mode == "verify_identity" {
		mode = "verify_full"
	}
	if mode != "required" && mode != "verify_ca" && mode != "verify_full" {
		return nil, fmt.Errorf("unsupported ZBX_DBTLSCONNECT value %q", db.env["ZBX_DBTLSCONNECT"])
	}

	config := &tls.Config{}
	caFile := db.env["ZBX_DBTLSCAFILE"]
	if caFile != "" || mode != "required" {
		roots, err := loadCertificatePool(caFile)
		if err != nil {
			return nil, err
		}
		config.RootCAs = roots
	}

	certFile := db.env["ZBX_DBTLSCERTFILE"]
	keyFile := db.env["ZBX_DBTLSKEYFILE"]
	if (certFile == "") != (keyFile == "") {
		return nil, fmt.Errorf("ZBX_DBTLSCERTFILE and ZBX_DBTLSKEYFILE must be set together")
	}
	if certFile != "" {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, fmt.Errorf("load database client certificate: %w", err)
		}
		config.Certificates = []tls.Certificate{cert}
	}

	switch mode {
	case "required":
		config.InsecureSkipVerify = true // Encryption is required, certificate verification is explicitly not requested
	case "verify_ca":
		config.InsecureSkipVerify = true // Verification is performed below without a DNS name
		config.VerifyConnection = verifyCertificateAuthority(config.RootCAs)
	case "verify_full":
		config.ServerName = strings.Trim(db.env["DB_SERVER_HOST"], "[]")
		if config.ServerName == "" {
			config.ServerName = "localhost"
		}
	}

	return config, nil
}

func loadCertificatePool(path string) (*x509.CertPool, error) {
	if path == "" {
		pool, err := x509.SystemCertPool()
		if err != nil {
			return nil, fmt.Errorf("load system certificate pool: %w", err)
		}
		return pool, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read database CA file %s: %w", path, err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(data) {
		return nil, fmt.Errorf("database CA file %s does not contain a certificate", path)
	}

	return pool, nil
}

func verifyCertificateAuthority(roots *x509.CertPool) func(tls.ConnectionState) error {
	return func(state tls.ConnectionState) error {
		if len(state.PeerCertificates) == 0 {
			return fmt.Errorf("database server did not provide a certificate")
		}

		intermediates := x509.NewCertPool()
		for _, cert := range state.PeerCertificates[1:] {
			intermediates.AddCert(cert)
		}
		_, err := state.PeerCertificates[0].Verify(x509.VerifyOptions{
			Roots:         roots,
			Intermediates: intermediates,
			KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		})

		return err
	}
}
