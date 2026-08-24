package bootstrap

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// DBTLSConfig contains TLS options for the entrypoint's database connection.
// Server/proxy and the PHP frontend expose different environment variables.
type DBTLSConfig struct {
	ConnectMode string
	CAFile      string
	CertFile    string
	KeyFile     string
}

// ServiceDBTLS reads the Zabbix server/proxy DBTLS settings.
func ServiceDBTLS(env Environment) DBTLSConfig {
	return DBTLSConfig{
		ConnectMode: env["ZBX_DBTLSCONNECT"],
		CAFile:      env["ZBX_DBTLSCAFILE"],
		CertFile:    env["ZBX_DBTLSCERTFILE"],
		KeyFile:     env["ZBX_DBTLSKEYFILE"],
	}
}

// FrontendDBTLS translates the PHP frontend DB TLS settings for Go clients.
func FrontendDBTLS(env Environment) DBTLSConfig {
	if !strings.EqualFold(env["ZBX_DB_ENCRYPTION"], "true") {
		return DBTLSConfig{}
	}

	settings := DBTLSConfig{
		ConnectMode: "required",
		CAFile:      env["ZBX_DB_CA_FILE"],
		CertFile:    env["ZBX_DB_CERT_FILE"],
		KeyFile:     env["ZBX_DB_KEY_FILE"],
	}
	if settings.CAFile != "" {
		settings.ConnectMode = "verify_ca"
		if strings.EqualFold(env["ZBX_DB_VERIFY_HOST"], "true") {
			settings.ConnectMode = "verify_full"
		}
	}

	return settings
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
