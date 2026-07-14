package bootstrap

// DBCredentials is a database username and password pair.
type DBCredentials struct {
	Username string
	Password string
}

// ApplyDatabaseCredentials exports user and password as ZBX_DB_USER /
// ZBX_DB_PASSWORD. When Zabbix is configured to fetch the credentials from
// Vault itself, both variables are removed instead so that they never reach
// the service environment.
func ApplyDatabaseCredentials(env Environment, user, password string) {
	if !useExplicitDatabaseCredentials(env) {
		delete(env, "ZBX_DB_USER")
		delete(env, "ZBX_DB_PASSWORD")
		return
	}

	env["ZBX_DB_USER"] = user
	env["ZBX_DB_PASSWORD"] = password
}

func useExplicitDatabaseCredentials(env Environment) bool {
	vault, vaultURL := env["ZBX_VAULT"], env["ZBX_VAULTURL"]
	if vault == "" && vaultURL == "" {
		return true
	}

	return vault != "" && vaultURL != "" && env["ZBX_VAULTDBPATH"] == ""
}
