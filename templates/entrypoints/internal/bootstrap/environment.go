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

// RequiredHomeDirectory returns the Zabbix home directory,
// verifying that it exists.
func RequiredHomeDirectory(env Environment) (string, error) {
	return requiredDirectory(env, "ZABBIX_USER_HOME_DIR")
}

// RequiredDirectories returns the Zabbix home and config directories
// (ZABBIX_USER_HOME_DIR and ZABBIX_CONF_DIR), verifying that both exist.
func RequiredDirectories(env Environment) (homeDir, configDir string, err error) {
	homeDir, err = RequiredHomeDirectory(env)
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

// ProcessTLSFiles moves TLS material from the listed variables into
// files under <home>/enc_internal, so that Zabbix reads certificates and
// keys from disk instead of the environment.
func ProcessTLSFiles(env Environment, homeDir string, variables ...string) error {
	directory := filepath.Join(homeDir, "enc_internal")
	for _, variable := range variables {
		fileVariable := variable + "FILE"

		if value := env[variable]; value != "" {
			path := filepath.Join(directory, fileVariable)

			if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
				return fmt.Errorf("write %s: %w", path, err)
			}

			env[fileVariable] = path
		}

		delete(env, variable)
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
