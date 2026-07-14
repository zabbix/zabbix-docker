package bootstrap

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
