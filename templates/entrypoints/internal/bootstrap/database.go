package bootstrap

import (
	"fmt"
	"path/filepath"
	"strings"
)

// DBCredentials is a database username and password pair.
type DBCredentials struct {
	Username string
	Password string
}

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

// ApplyDBCredentials exports user and password as ZBX_DB_USER /
// ZBX_DB_PASSWORD. When Zabbix is configured to fetch the credentials from
// Vault itself, both variables are removed instead so that they never reach
// the service environment.
func ApplyDBCredentials(env Environment, user, password string) {
	if !hasExplicitDBCredentials(env) {
		delete(env, "ZBX_DB_USER")
		delete(env, "ZBX_DB_PASSWORD")
		return
	}

	env["ZBX_DB_USER"] = user
	env["ZBX_DB_PASSWORD"] = password
}

func hasExplicitDBCredentials(env Environment) bool {
	vault, vaultURL := env["ZBX_VAULT"], env["ZBX_VAULTURL"]
	if vault == "" && vaultURL == "" {
		return true
	}

	return vault != "" && vaultURL != "" && env["ZBX_VAULTDBPATH"] == ""
}
