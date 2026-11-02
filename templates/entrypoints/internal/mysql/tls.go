package mysql

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

func (db *DB) tlsConfig() (*tls.Config, error) {
	mode := strings.ToLower(db.tls.ConnectMode)
	if mode == "" {
		return nil, nil
	}
	if mode != "required" && mode != "verify_ca" && mode != "verify_full" {
		return nil, fmt.Errorf("unsupported database TLS connection mode %q", db.tls.ConnectMode)
	}

	config := &tls.Config{}
	caFile := db.tls.CAFile
	if caFile != "" || mode != "required" {
		roots, err := loadCertificatePool(caFile)
		if err != nil {
			return nil, err
		}
		config.RootCAs = roots
	}

	certFile := db.tls.CertFile
	keyFile := db.tls.KeyFile
	if (certFile == "") != (keyFile == "") {
		return nil, errors.New("ZBX_DBTLSCERTFILE and ZBX_DBTLSKEYFILE must be set together")
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
		config.VerifyConnection = func(state tls.ConnectionState) error {
			return bootstrap.VerifyCertificateChain(state.PeerCertificates, config.RootCAs)
		}
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
