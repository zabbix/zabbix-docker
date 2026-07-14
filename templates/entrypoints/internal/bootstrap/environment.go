package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Environment map[string]string

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

func (env Environment) ValueOrDefault(name, defaultValue string) string {
	value, found := env[name]
	if !found {
		return defaultValue
	}
	return value
}

func (env Environment) SetDefault(name, defaultValue string) {
	if _, found := env[name]; !found {
		env[name] = defaultValue
	}
}

func (env Environment) SetDefaultNonEmpty(name, defaultValue string) {
	if env[name] == "" {
		env[name] = defaultValue
	}
}

func (env Environment) ValueOrDefaultNonEmpty(name, defaultValue string) string {
	if value := env[name]; value != "" {
		return value
	}
	return defaultValue
}

func RequiredHomeDirectory(env Environment) (string, error) {
	return requiredDirectory(env, "ZABBIX_USER_HOME_DIR")
}

func RequiredDirectories(env Environment) (homeDirectory, configDirectory string, err error) {
	homeDirectory, err = RequiredHomeDirectory(env)
	if err != nil {
		return "", "", err
	}

	configDirectory, err = requiredDirectory(env, "ZABBIX_CONF_DIR")
	if err != nil {
		return "", "", err
	}

	return homeDirectory, configDirectory, nil
}

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

func FileEnv(env Environment, name, defaultValue string) error {
	fileName := name + "_FILE"
	value := env[name]
	secretFile := env[fileName]
	if value != "" && secretFile != "" {
		return fmt.Errorf("both variables %s and %s are set (but are exclusive)", name, fileName)
	}

	value = defaultValue
	if env[name] != "" {
		value = env[name]
		LogInfo("** Using %s variable from ENV", name)
	} else if secretFile != "" {
		data, err := os.ReadFile(secretFile)
		if err != nil {
			return fmt.Errorf("secret file %q is not found: %w", secretFile, err)
		}
		value = strings.TrimRight(string(data), "\n")
		LogInfo("** Using %s variable from secret file", name)
	}

	env[name] = value
	delete(env, fileName)
	return nil
}

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

func ProcessEncryptionFiles(env Environment, homeDirectory string, variables ...string) error {
	directory := filepath.Join(homeDirectory, "enc_internal")
	for _, variable := range variables {
		if err := ProcessFileFromEnvironment(env, directory, variable); err != nil {
			return err
		}
	}

	return nil
}

func ClearPrivateEnvironment(env Environment, prefixes ...string) {
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
