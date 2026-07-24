//go:build !windows

package hooks

import (
	"os"
	"path/filepath"
	"strings"
)

func hookCommand(path string, mode os.FileMode) ([]string, bool) {
	if strings.EqualFold(filepath.Ext(path), ".sh") {
		return []string{"/bin/sh", path}, true
	}
	if mode.Perm()&0o111 != 0 {
		return []string{path}, true
	}

	return nil, false
}
