package hooks

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

const directoryName = "entrypoint.d"

func Run(env bootstrap.Environment) error {
	homeDirectory, err := bootstrap.RequiredHomeDirectory(env)
	if err != nil {
		return err
	}

	directory := filepath.Join(homeDirectory, directoryName)
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

		args, supported := command(path)
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
