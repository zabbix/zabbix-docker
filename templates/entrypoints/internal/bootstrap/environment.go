//go:build windows

// Package bootstrap provides the shared building blocks of the container
// entrypoints: process environment handling, Zabbix configuration file
// updates, logging and the final hand-off to the service binary.
package bootstrap

import (
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

// RequiredDirectories returns the Zabbix home and config directories
// (ZABBIX_USER_HOME_DIR and ZABBIX_CONF_DIR), verifying that both exist.
func RequiredDirectories(env Environment) (homeDir, configDir string, err error) {
	homeDir, err = requiredDirectory(env, "ZABBIX_USER_HOME_DIR")
	if err != nil {
		return "", "", err
	}

	configDir, err = requiredDirectory(env, "ZABBIX_CONF_DIR")
	if err != nil {
		return "", "", err
	}

	return homeDir, configDir, nil
}

// requiredDirectory returns directory from env variable.
func requiredDirectory(env Environment, name string) (string, error) {
	directory := env[name]
	if directory == "" {
		return "", fmt.Errorf("%s must be set", name)
	}

	info, err := os.Stat(directory)
	if err != nil {
		return "", fmt.Errorf("access %s directory %s: %w", name, directory, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s path %s is not a directory", name, directory)
	}

	return directory, nil
}

// ProcessFileFromEnvironment persists the value of variable into a file
// under directory and points the corresponding "<variable>FILE" variable at
// it. The plain variable is removed from the environment.
func ProcessFileFromEnvironment(env Environment, directory, variable string) error {
	fileVariable := variable + "FILE"

	if value := env[variable]; value != "" {
		path := filepath.Join(directory, fileVariable)
		if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
		env[fileVariable] = path
	}
	delete(env, variable)

	return nil
}

// ProcessTLSFiles persists inline TLS material under <home>/enc_internal,
// resolves relative TLS file names against <home>/enc and updates configPath.
func ProcessTLSFiles(env Environment, homeDir, configPath string, parameters ...string) error {
	internalDirectory := filepath.Join(homeDir, "enc_internal")

	for _, parameter := range parameters {
		variable := "ZBX_" + strings.ToUpper(strings.TrimSuffix(parameter, "File"))
		fileVariable := variable + "FILE"
		if value := env[variable]; value != "" {
			path := filepath.Join(internalDirectory, parameter)
			if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
				return fmt.Errorf("write %s: %w", path, err)
			}
			env[fileVariable] = path
		}
		delete(env, variable)

		if value := env[fileVariable]; value != "" && !filepath.IsAbs(value) {
			env[fileVariable] = filepath.Join(homeDir, "enc", value)
		}

		if err := UpdateConfigValue(configPath, parameter, env[fileVariable]); err != nil {
			return err
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
