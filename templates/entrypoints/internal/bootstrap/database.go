package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	envDBTLSCAFile   = "ZBX_DBTLSCAFILE"
	envDBTLSCertFile = "ZBX_DBTLSCERTFILE"
	envDBTLSKeyFile  = "ZBX_DBTLSKEYFILE"
)

var dbTLSRuntimeDir = "/run/zabbix/secrets/database"

var dbTLSVariables = []struct {
	name     string
	fileName string
}{
	{"ZBX_DBTLSCA", "ca.pem"},
	{"ZBX_DBTLSCERT", "cert.pem"},
	{"ZBX_DBTLSKEY", "key.pem"},
}

// PrepareDBTLSFiles always moves inline material out of the environment
// (ClearPrivateEnv keeps ZBX_*); *FILE paths are resolved only if resolvePaths.
func PrepareDBTLSFiles(env Environment, resolvePaths bool) error {
	inline := false
	for _, variable := range dbTLSVariables {
		fileVariable := variable.name + "FILE"
		if env[variable.name] != "" && env[fileVariable] != "" {
			return fmt.Errorf("both variables %s and %s are set (but are exclusive)", variable.name, fileVariable)
		}
		inline = inline || env[variable.name] != ""
	}
	if inline {
		if err := os.MkdirAll(dbTLSRuntimeDir, 0o700); err != nil {
			return fmt.Errorf("create database TLS directory: %w", err)
		}
	}

	for _, variable := range dbTLSVariables {
		if env[variable.name] == "" && !resolvePaths {
			delete(env, variable.name)
			continue
		}
		path := filepath.Join(dbTLSRuntimeDir, variable.fileName)
		if err := processTLSFileTo(env, SecretsDir, path, variable.name); err != nil {
			return err
		}
	}
	return nil
}

// DBTLSConfig contains TLS options for the entrypoint's database connection.
// Server/proxy and the PHP frontend expose different environment variables.
type DBTLSConfig struct {
	ConnectMode string
	CAFile      string
	CertFile    string
	KeyFile     string
}

// ResolveServiceDBTLS resolves database TLS file names for Zabbix server and
// proxy images.
func ResolveServiceDBTLS(env Environment) (DBTLSConfig, error) {
	mode := env["ZBX_DBTLSCONNECT"]
	if err := PrepareDBTLSFiles(env, mode != ""); err != nil {
		return DBTLSConfig{}, err
	}
	return dbTLSConfig(env, mode), nil
}

// ResolveFrontendDBTLS resolves database TLS file names for frontend images.
func ResolveFrontendDBTLS(env Environment) (DBTLSConfig, error) {
	// The mode below depends on the CA file.
	encrypted := strings.EqualFold(env["ZBX_DB_ENCRYPTION"], "true")
	// Frontend file aliases have historically been resolved regardless of
	// whether database encryption is enabled.
	if err := PrepareDBTLSFiles(env, true); err != nil {
		return DBTLSConfig{}, err
	}
	if !encrypted {
		return DBTLSConfig{}, nil
	}

	mode := "required"
	if env[envDBTLSCAFile] != "" {
		mode = "verify_ca"
		if strings.EqualFold(env["ZBX_DB_VERIFY_HOST"], "true") {
			mode = "verify_full"
		}
	}

	return dbTLSConfig(env, mode), nil
}

func dbTLSConfig(env Environment, mode string) DBTLSConfig {
	return DBTLSConfig{
		ConnectMode: mode,
		CAFile:      env[envDBTLSCAFile],
		CertFile:    env[envDBTLSCertFile],
		KeyFile:     env[envDBTLSKeyFile],
	}
}

// ResolveDBPort returns DB_SERVER_PORT or defaultPort and validates that the
// value is a numeric port. The resolved value is written back to env so
// bootstrap clients and the final Zabbix process use the same setting.
func ResolveDBPort(env Environment, defaultPort string) (string, error) {
	value := env.ValueOrDefaultNonEmpty("DB_SERVER_PORT", defaultPort)
	if _, err := strconv.ParseUint(value, 10, 16); err != nil {
		return "", fmt.Errorf("invalid DB_SERVER_PORT %q: %w", value, err)
	}

	env["DB_SERVER_PORT"] = value

	return value, nil
}

// RunAdditionalSQLScripts executes every *.sql file from the dbscripts
// directory under the Zabbix home, in file name order.
func RunAdditionalSQLScripts(env Environment, execute func(path string) error) error {
	homeDir, err := RequiredHomeDir(env)
	if err != nil {
		return err
	}

	scripts, err := filepath.Glob(filepath.Join(homeDir, "dbscripts", "*.sql"))
	if err != nil {
		return fmt.Errorf("find additional SQL scripts: %w", err)
	}
	for _, script := range scripts {
		LogInfo("** Processing additional '%s' SQL script", script)
		if err := execute(script); err != nil {
			return err
		}
	}

	return nil
}
