package bootstrap

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/testutil"
)

func TestPrepareCertDir(t *testing.T) {
	sourceDir := t.TempDir()
	targetDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(sourceDir, "ca.pem"), []byte("certificate"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(targetDir, "stale"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(targetDir, "stale", "file"), nil, 0o600); err != nil {
		t.Fatal(err)
	}

	testutil.CaptureStderr(t, func() {
		if err := PrepareCertDir(sourceDir, targetDir); err != nil {
			t.Fatal(err)
		}
	})

	if _, err := os.Stat(filepath.Join(targetDir, "stale")); !os.IsNotExist(err) {
		t.Fatalf("stale directory was not removed: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(targetDir, "ca.pem"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "certificate" {
		t.Fatalf("certificate = %q", data)
	}
}
