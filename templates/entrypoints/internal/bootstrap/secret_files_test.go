package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveExternalSecretFile(t *testing.T) {
	root := t.TempDir()
	oldRoot := externalSecretsRoot
	externalSecretsRoot = root
	t.Cleanup(func() { externalSecretsRoot = oldRoot })

	certDir := filepath.Join(root, "ssl", "certs")
	if err := os.MkdirAll(certDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cert := filepath.Join(certDir, "client.crt")
	if err := os.WriteFile(cert, []byte("certificate"), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := ResolveExternalSecretFile("ssl_cert_file", "client.crt", "ssl/certs")
	if err != nil {
		t.Fatal(err)
	}
	if got != cert {
		t.Fatalf("resolved path = %q, want %q", got, cert)
	}

	if _, err := ResolveExternalSecretFile("ssl_cert_file", "missing.crt", "ssl/certs"); err == nil ||
		!strings.Contains(err.Error(), "was not found") {
		t.Fatalf("missing file error = %v", err)
	}
	if _, err := ResolveExternalSecretFile("ssl_cert_file", "../client.crt", "ssl/certs"); err == nil {
		t.Fatal("escaping relative path was accepted")
	}
}
