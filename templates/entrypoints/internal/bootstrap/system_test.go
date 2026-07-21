package bootstrap

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRehashCertificateDirectoryLogsNonDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ca.pem")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}

	output := captureStderr(t, func() {
		RehashCertificateDirectory(path)
	})
	if !strings.Contains(output, "not a directory") || !strings.Contains(output, path) {
		t.Fatalf("warning = %q", output)
	}
}

func TestRehashCertificateDirectoryIgnoresMissingDirectory(t *testing.T) {
	output := captureStderr(t, func() {
		RehashCertificateDirectory(filepath.Join(t.TempDir(), "missing"))
	})
	if output != "" {
		t.Fatalf("warning = %q", output)
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
