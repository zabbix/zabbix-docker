package bootstrap

import (
	"io"
	"os"
	"path/filepath"
	"testing"
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

	captureStderr(t, func() {
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

func captureStderr(t *testing.T, run func()) string {
	t.Helper()

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	stderr := os.Stderr
	os.Stderr = writer
	defer func() {
		os.Stderr = stderr
	}()

	run()
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}

	return string(output)
}
