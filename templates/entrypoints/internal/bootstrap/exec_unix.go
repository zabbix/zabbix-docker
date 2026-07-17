//go:build !windows

package bootstrap

import (
	"os/exec"
	"strings"
	"syscall"
)

// executeProcess replaces the current process with args, resolving the binary via
// PATH when necessary.
func executeProcess(args []string, env Environment) error {
	path := args[0]
	if !strings.ContainsRune(path, '/') {
		resolvedPath, err := exec.LookPath(path)
		if err != nil {
			return err
		}
		path = resolvedPath
	}
	return syscall.Exec(path, args, env.List())
}
