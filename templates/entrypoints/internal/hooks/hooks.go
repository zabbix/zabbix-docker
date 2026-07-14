// Package hooks runs user-provided scripts from the entrypoint.d directory
// before the service starts.
package hooks

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

const directoryName = "entrypoint.d"

// Run executes the files from <home>/entrypoint.d in file name order: *.sh
// scripts via the shell, executable files directly; everything else is
// skipped. The first failing hook aborts the entrypoint.
func Run(env bootstrap.Environment) error {
	homeDir, err := bootstrap.RequiredHomeDirectory(env)
	if err != nil {
		return err
	}

	directory := filepath.Join(homeDir, directoryName)
	entries, err := os.ReadDir(directory)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read entrypoint hooks directory %s: %w", directory, err)
	}

	for _, entry := range entries {
		path := filepath.Join(directory, entry.Name())
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("inspect entrypoint hook %s: %w", path, err)
		}
		if !info.Mode().IsRegular() {
			continue
		}

		args, supported := command(path, info.Mode())
		if !supported {
			continue
		}

		bootstrap.LogInfo("** Running entrypoint hook: %s", path)

		hook := exec.Command(args[0], args[1:]...)
		hook.Env = env.List()
		hook.Stdin = os.Stdin
		hook.Stdout = os.Stdout
		hook.Stderr = os.Stderr
		if err := hook.Run(); err != nil {
			return fmt.Errorf("entrypoint hook %s failed: %w", path, err)
		}
	}

	return nil
}
