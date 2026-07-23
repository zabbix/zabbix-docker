package bootstrap

import (
	"fmt"
	"path/filepath"
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

// RunAdditionalSQLScripts executes every *.sql file from the dbscripts
// directory under the Zabbix home, in file name order.
func RunAdditionalSQLScripts(env Environment, execute func(path string) error) error {
	homeDir, err := RequiredHomeDirectory(env)
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
