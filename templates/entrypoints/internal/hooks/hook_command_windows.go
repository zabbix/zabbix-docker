package hooks

import (
	"os"
	"path/filepath"
	"strings"
)

func hookCommand(path string, _ os.FileMode) ([]string, bool) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".ps1":
		return []string{"pwsh.exe", "-NoLogo", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-File", path}, true
	case ".cmd", ".bat":
		return []string{"cmd.exe", "/D", "/S", "/C", path}, true
	default:
		return nil, false
	}
}
