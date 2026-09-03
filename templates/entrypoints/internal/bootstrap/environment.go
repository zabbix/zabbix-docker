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

// SetDefault stores defaultValue unless the variable is already present.
func (env Environment) SetDefault(name, defaultValue string) {
	if _, found := env[name]; !found {
		env[name] = defaultValue
	}
}

// SetDefaultNonEmpty stores defaultValue unless the variable already has a
// non-empty value.
func (env Environment) SetDefaultNonEmpty(name, defaultValue string) {
	if env[name] == "" {
		env[name] = defaultValue
	}
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

// ResolveSecretEnv resolves the NAME / NAME_FILE pair following the Docker
// secrets convention. The value is read from NAME_FILE when NAME is empty;
// setting both is an error. The *_FILE variable is removed from the
// environment.
func ResolveSecretEnv(env Environment, name string) error {
	fileName := name + "_FILE"
	value := env[name]
	secretFile := env[fileName]
	if value != "" && secretFile != "" {
		return fmt.Errorf("both variables %s and %s are set (but are exclusive)", name, fileName)
	}

	if value != "" {
		LogInfo("** Using %s variable from ENV", name)
	} else if secretFile != "" {
		data, err := os.ReadFile(secretFile)
		if err != nil {
			return fmt.Errorf("secret file %q is not found: %w", secretFile, err)
		}
		value = strings.TrimRight(string(data), "\r\n")
		LogInfo("** Using %s variable from secret file", name)
	}

	env[name] = value
	delete(env, fileName)
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

// ClearPrivateEnv removes configuration variables that the service
// should not inherit, by default everything with the ZABBIX_, DB_, MYSQL_
// and POSTGRES_ prefixes. ZBX_CLEAR_ENV=false disables the cleanup.
func ClearPrivateEnv(env Environment, prefixes ...string) {
	if env["ZBX_CLEAR_ENV"] == "false" {
		return
	}

	if len(prefixes) == 0 {
		prefixes = []string{"ZABBIX_", "DB_", "MYSQL_", "POSTGRES_"}
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
