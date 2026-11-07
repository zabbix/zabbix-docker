//go:build !windows

package bootstrap

import (
	"os"
	"path/filepath"
	"testing"
)

func TestChownGeneratedSecretTreeSkipsSymlinks(t *testing.T) {
	directory := t.TempDir()
	if err := os.MkdirAll(filepath.Join(directory, "providers", "telemetry-0"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "providers", "telemetry-0", "password"), []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(t.TempDir(), "missing"), filepath.Join(directory, "dangling")); err != nil {
		t.Fatal(err)
	}

	root, err := os.OpenRoot(directory)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()

	if err := chownGeneratedSecretTree(root, os.Getuid(), os.Getgid()); err != nil {
		t.Fatal(err)
	}
}
