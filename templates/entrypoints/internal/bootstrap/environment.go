//go:build windows

// Package bootstrap provides the shared building blocks of the container
// entrypoints: process environment handling, Zabbix configuration file
// updates, logging and the final hand-off to the service binary.
package bootstrap

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Environment is a mutable set of environment variables keyed by name.
// The entrypoint modifies it while preparing a service and passes the
// result to the final process.
type Environment map[string]string

// NewEnvironment parses "NAME=value" pairs as returned by os.Environ.
func NewEnvironment(values []string) Environment {
	env := make(Environment, len(values))
	for _, item := range values {
		name, value, found := strings.Cut(item, "=")
		if found {
			env[name] = value
		}
	}
	return env
}

// List returns the variables as sorted "NAME=value" pairs suitable for exec.
func (env Environment) List() []string {
	names := make([]string, 0, len(env))
	for name := range env {
		names = append(names, name)
	}
	sort.Strings(names)

	values := make([]string, 0, len(names))
	for _, name := range names {
		values = append(values, name+"="+env[name])
	}
	return values
}

// ValueOrDefault returns the value of name, or defaultValue when the
// variable is not present. An empty value counts as present; use
// ValueOrDefaultNonEmpty to treat it as missing.
func (env Environment) ValueOrDefault(name, defaultValue string) string {
	value, found := env[name]
	if !found {
		return defaultValue
	}
	return value
}

// ValueOrDefaultNonEmpty returns the value of name, or defaultValue when
// the variable is missing or empty.
func (env Environment) ValueOrDefaultNonEmpty(name, defaultValue string) string {
	if value := env[name]; value != "" {
		return value
	}
	return defaultValue
}

// RequiredHomeDir returns the Zabbix home directory,
// verifying that it exists.
func RequiredHomeDir(env Environment) (string, error) {
	dir := env["ZABBIX_USER_HOME_DIR"]
	if err := validateDir(dir); err != nil {
		return "", fmt.Errorf("ZABBIX_USER_HOME_DIR: %w", err)
	}

	return dir, nil
}

// RequiredConfigDir returns the Zabbix config directory,
// verifying that it exists.
func RequiredConfigDir(env Environment) (string, error) {
	dir := env["ZABBIX_CONF_DIR"]
	if err := validateDir(dir); err != nil {
		return "", fmt.Errorf("ZABBIX_CONF_DIR: %w", err)
	}

	return dir, nil
}

// CommonDirs returns the Zabbix home and config directories
// (ZABBIX_USER_HOME_DIR and ZABBIX_CONF_DIR), verifying that both exist.
func CommonDirs(env Environment) (homeDir, configDir string, err error) {
	homeDir, err = RequiredHomeDir(env)
	if err != nil {
		return "", "", err
	}

	configDir, err = RequiredConfigDir(env)
	if err != nil {
		return "", "", err
	}

	return homeDir, configDir, nil
}

// validateDir verifies that dir is an existing directory.
func validateDir(dir string) error {
	if dir == "" {
		return errors.New("must be set")
	}

	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("access %q: %w", dir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%q is not a directory", dir)
	}

	return nil
}

// ProcessTLSFiles moves inline TLS material from the listed variables into
// files under <home>/enc_internal. Existing relative TLS file paths are
// resolved against <home>/enc, while absolute paths are preserved unchanged.
func ProcessTLSFiles(env Environment, homeDir string, variables ...string) error {
	internalDir := filepath.Join(homeDir, "enc_internal")
	volumeDir := filepath.Join(homeDir, "enc")

	for _, variable := range variables {
		fileVariable := variable + "FILE"

		if value := env[variable]; value != "" {
			path := filepath.Join(internalDir, fileVariable)

			if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
				return fmt.Errorf("write %s: %w", path, err)
			}

			env[fileVariable] = path
		}

		delete(env, variable)

		if path := env[fileVariable]; path != "" && !filepath.IsAbs(path) {
			env[fileVariable] = filepath.Join(volumeDir, path)
		}
	}

	return nil
}

// ClearPrivateEnv removes variables with the supplied prefixes that the
// service should not inherit. ZBX_CLEAR_ENV=false disables the cleanup.
func ClearPrivateEnv(env Environment, prefixes ...string) {
	if env["ZBX_CLEAR_ENV"] == "false" {
		return
	}
	for name := range env {
		if hasAnyPrefix(name, prefixes) {
			delete(env, name)
		}
	}
}

func hasAnyPrefix(value string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}
