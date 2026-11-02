package bootstrap

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	envDBTLSCAFile   = "ZBX_DBTLSCAFILE"
	envDBTLSCertFile = "ZBX_DBTLSCERTFILE"
	envDBTLSKeyFile  = "ZBX_DBTLSKEYFILE"
)

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
	return resolveDBTLS(env, env["ZBX_DBTLSCONNECT"])
}

// ResolveFrontendDBTLS resolves database TLS file names for frontend images.
func ResolveFrontendDBTLS(env Environment) (DBTLSConfig, error) {
	if !strings.EqualFold(env["ZBX_DB_ENCRYPTION"], "true") {
		return DBTLSConfig{}, nil
	}

	mode := "required"
	if env[envDBTLSCAFile] != "" {
		mode = "verify_ca"
		if strings.EqualFold(env["ZBX_DB_VERIFY_HOST"], "true") {
			mode = "verify_full"
		}
	}

	return resolveDBTLS(env, mode)
}

func resolveDBTLS(env Environment, mode string) (DBTLSConfig, error) {
	if mode != "" {
		for _, name := range []string{envDBTLSCAFile, envDBTLSCertFile, envDBTLSKeyFile} {
			if err := ResolveFileEnv(env, name, SecretsDir); err != nil {
				return DBTLSConfig{}, err
			}
		}
	}

	return DBTLSConfig{
		ConnectMode: mode,
		CAFile:      env[envDBTLSCAFile],
		CertFile:    env[envDBTLSCertFile],
		KeyFile:     env[envDBTLSKeyFile],
	}, nil
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
