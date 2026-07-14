//go:build !windows

package bootstrap

import (
	"os/exec"
	"strings"
	"syscall"
)

func Exec(args []string, env Environment) error {
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
