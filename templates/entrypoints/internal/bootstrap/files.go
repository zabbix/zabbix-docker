//go:build windows

package bootstrap

import (
	"os"
	"path/filepath"
)

// WriteFilePreservingMode atomically replaces a file while preserving its
// permission bits.
func WriteFilePreservingMode(path string, data []byte) error {
	resolvedPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return err
	}

	info, err := os.Stat(resolvedPath)
	if err != nil {
		return err
	}

	temporary, err := os.CreateTemp(filepath.Dir(resolvedPath), "."+filepath.Base(resolvedPath)+".tmp-*")
	if err != nil {
		return err
	}
	defer os.Remove(temporary.Name())
	defer temporary.Close()

	if err := temporary.Chmod(info.Mode().Perm()); err != nil {
		return err
	}
	if _, err := temporary.Write(data); err != nil {
		return err
	}
	if err := temporary.Sync(); err != nil {
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}

	return os.Rename(temporary.Name(), resolvedPath)
}
