package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSecretPlanDoesNotModifyEnvironment(t *testing.T) {
	const variable = "ZBX_HISTORYPROVIDER_0_PASSWORD"
	env := Environment{
		"ZBX_HISTORYPROVIDER_0": `{"provider":"clickhouse"}`,
		variable:                "secret",
		"ZABBIX_USER_HOME_DIR":  t.TempDir(),
	}
	if _, _, _, err := secretPlan(env, SecretsProxyMySQL, SecretsService); err != nil {
		t.Fatal(err)
	}
	if env[variable] != "secret" {
		t.Fatalf("secretPlan modified %s", variable)
	}
}

func TestResolveSecretsDropsUnsupportedProxyProviderSources(t *testing.T) {
	_, _, _ = useSecretRoots(t)
	const variable = "ZBX_HISTORYPROVIDER_0_PASSWORD"
	env := Environment{
		"ZBX_HISTORYPROVIDER_0": `{"provider":"clickhouse"}`,
		variable:                "secret",
		"ZABBIX_USER_HOME_DIR":  t.TempDir(),
	}
	if err := ResolveSecrets(env, SecretsProxyMySQL, SecretsInitDB); err != nil {
		t.Fatal(err)
	}
	if _, exists := env[variable]; exists {
		t.Fatalf("ResolveSecrets left unsupported %s", variable)
	}
}

func TestTelemetryProviderUsesVaultParsesProviderJSON(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want bool
	}{
		{name: "vault path", raw: `{"provider":"clickhouse","vault_path":"secret/telemetry"}`, want: true},
		{name: "single quotes", raw: `{'provider':'clickhouse','vault_path':'secret/telemetry'}`, want: true},
		{name: "empty vault path", raw: `{"provider":"clickhouse","vault_path":""}`},
		{name: "vault path in value", raw: `{"description":"vault_path"}`},
		{name: "non-string vault path", raw: `{"vault_path":1}`},
		{name: "invalid JSON", raw: `{"vault_path":`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env := Environment{"ZBX_TELEMETRYPROVIDER_0": test.raw}
			if got := telemetryProviderUsesVault(env); got != test.want {
				t.Fatalf("telemetryProviderUsesVault() = %t, want %t", got, test.want)
			}
		})
	}
}

func TestSecretPlanDisablesLocalDatabaseCredentialsWithVault(t *testing.T) {
	tests := []struct {
		name      string
		component SecretComponent
		want      map[string]bool
	}{
		{
			name:      "MySQL",
			component: SecretsServerMySQL,
			want: map[string]bool{
				"MYSQL_USER":          false,
				"MYSQL_PASSWORD":      false,
				"MYSQL_ROOT_USER":     true,
				"MYSQL_ROOT_PASSWORD": true,
			},
		},
		{
			name:      "PostgreSQL",
			component: SecretsServerPgSQL,
			want: map[string]bool{
				"POSTGRES_USER":     false,
				"POSTGRES_PASSWORD": false,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env := Environment{"ZBX_VAULTDBPATH": "secret/database"}
			_, values, _, err := secretPlan(env, test.component, SecretsInitDB)
			if err != nil {
				t.Fatal(err)
			}
			got := make(map[string]bool)
			for _, secret := range values {
				if _, expected := test.want[secret.name]; expected {
					got[secret.name] = secret.active
				}
			}
			for name, want := range test.want {
				active, found := got[name]
				if !found || active != want {
					t.Errorf("%s active = %t, found = %t; want %t", name, active, found, want)
				}
			}
		})
	}
}

func TestResolveSecretsIgnoresLocalDatabaseFilesWithVault(t *testing.T) {
	tests := []struct {
		name      string
		component SecretComponent
		file      string
	}{
		{name: "MySQL user", component: SecretsServerMySQL, file: "MYSQL_USER_FILE"},
		{name: "MySQL password", component: SecretsServerMySQL, file: "MYSQL_PASSWORD_FILE"},
		{name: "PostgreSQL user", component: SecretsServerPgSQL, file: "POSTGRES_USER_FILE"},
		{name: "PostgreSQL password", component: SecretsServerPgSQL, file: "POSTGRES_PASSWORD_FILE"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, _, _ = useSecretRoots(t)
			env := Environment{
				"ZABBIX_USER_HOME_DIR": t.TempDir(),
				"ZBX_VAULTDBPATH":      "secret/database",
				test.file:              "missing",
			}
			if err := ResolveSecrets(env, test.component, SecretsInitDB); err != nil {
				t.Fatal(err)
			}
			if _, exists := env[test.file]; exists {
				t.Fatalf("inactive %s was not removed", test.file)
			}
		})
	}
}

func TestResolveSecretsScopesHistoryProviderSecrets(t *testing.T) {
	const provider = `{"provider":"clickhouse","url":"https://clickhouse:8443","db":"zabbix"}`
	tests := []struct {
		name      string
		component SecretComponent
		variable  string
		wantError bool
	}{
		{
			name: "web ignores TLS file", component: SecretsWebMySQL,
			variable: "ZBX_HISTORYPROVIDER_0_SSL_CERTFILE",
		},
		{
			name: "server validates TLS file", component: SecretsServerMySQL,
			variable: "ZBX_HISTORYPROVIDER_0_SSL_CERTFILE", wantError: true,
		},
		{
			name: "web validates username file", component: SecretsWebMySQL,
			variable: "ZBX_HISTORYPROVIDER_0_USERNAME_FILE", wantError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, _, _ = useSecretRoots(t)
			env := Environment{
				"ZABBIX_USER_HOME_DIR":  t.TempDir(),
				"ZBX_HISTORYPROVIDER_0": provider,
				test.variable:           "missing",
			}
			err := ResolveSecrets(env, test.component, SecretsService)
			if test.wantError {
				if err == nil || !strings.Contains(err.Error(), test.variable) {
					t.Fatalf("ResolveSecrets() error = %v, want %s context", err, test.variable)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if _, exists := env[test.variable]; exists {
				t.Fatalf("unused %s was not removed", test.variable)
			}
		})
	}
}

func TestResolveSecretsDiscardsProviderLocationInputs(t *testing.T) {
	const provider = `{"provider":"clickhouse","url":"https://clickhouse:8443","db":"zabbix"}`
	for _, family := range []string{"HISTORY", "TELEMETRY"} {
		for _, secret := range []string{"CA", "CERT", "KEY"} {
			variable := "ZBX_" + family + "PROVIDER_0_SSL_" + secret + "LOCATION"
			t.Run(family+" "+secret, func(t *testing.T) {
				_, _, _ = useSecretRoots(t)
				env := Environment{
					"ZABBIX_USER_HOME_DIR":         t.TempDir(),
					"ZBX_" + family + "PROVIDER_0": provider,
					variable:                       "/user/supplied",
				}
				if err := ResolveSecrets(env, SecretsServerMySQL, SecretsService); err != nil {
					t.Fatal(err)
				}
				if _, exists := env[variable]; exists {
					t.Fatalf("internal output variable %s was not removed", variable)
				}
			})
		}
	}
}

func TestResolveSecretsGeneratesTelemetryProviderCALocation(t *testing.T) {
	for name, component := range map[string]SecretComponent{
		"server": SecretsServerMySQL,
		"proxy":  SecretsProxyMySQL,
	} {
		t.Run(name, func(t *testing.T) {
			external, internal, _ := useSecretRoots(t)
			writeTestSecret(t, filepath.Join(external, "providers", "telemetry-0", "ca.pem"), "certificate")
			env := Environment{
				"ZABBIX_USER_HOME_DIR":                   t.TempDir(),
				"ZBX_TELEMETRYPROVIDER_0":                `{"provider":"clickhouse","url":"https://clickhouse:8443","db":"zabbix"}`,
				"ZBX_TELEMETRYPROVIDER_0_SSL_CAFILE":     "ca.pem",
				"ZBX_TELEMETRYPROVIDER_0_SSL_CALOCATION": "/user/supplied",
			}
			if err := ResolveSecrets(env, component, SecretsService); err != nil {
				t.Fatal(err)
			}
			want := filepath.Join(internal, "providers", "telemetry-0", "ca")
			if env["ZBX_TELEMETRYPROVIDER_0_SSL_CALOCATION"] != want {
				t.Fatalf("CA location = %q, want %q", env["ZBX_TELEMETRYPROVIDER_0_SSL_CALOCATION"], want)
			}
		})
	}
}

func TestResolveSecretsNamesInlineTelemetryCAForRehash(t *testing.T) {
	_, internal, _ := useSecretRoots(t)
	env := Environment{
		"ZABBIX_USER_HOME_DIR":           t.TempDir(),
		"ZBX_TELEMETRYPROVIDER_0":        `{"provider":"clickhouse","url":"https://clickhouse:8443","db":"zabbix"}`,
		"ZBX_TELEMETRYPROVIDER_0_SSL_CA": "certificate",
	}
	if err := ResolveSecrets(env, SecretsProxyMySQL, SecretsService); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(internal, "providers", "telemetry-0", "ca", "ZBX_TELEMETRYPROVIDER_0_SSL_CAFILE.pem")
	if data, err := os.ReadFile(path); err != nil || string(data) != "certificate" {
		t.Fatalf("generated CA=%q, err=%v", data, err)
	}
}

func TestResolveSecretsKeepsWebTelemetryCAAsFile(t *testing.T) {
	external, internal, _ := useSecretRoots(t)
	path := filepath.Join(external, "providers", "telemetry-0", "ca.pem")
	writeTestSecret(t, path, "certificate")
	env := Environment{
		"ZABBIX_USER_HOME_DIR":               t.TempDir(),
		"ZBX_TELEMETRYPROVIDER_0":            `{"provider":"clickhouse","url":"https://clickhouse:8443","db":"zabbix"}`,
		"ZBX_TELEMETRYPROVIDER_0_SSL_CAFILE": "ca.pem",
	}
	if err := ResolveSecrets(env, SecretsWebMySQL, SecretsService); err != nil {
		t.Fatal(err)
	}
	if env["ZBX_TELEMETRYPROVIDER_0_SSL_CAFILE"] != path {
		t.Fatalf("CA file=%q, want %q", env["ZBX_TELEMETRYPROVIDER_0_SSL_CAFILE"], path)
	}
	if env["ZBX_TELEMETRYPROVIDER_0_SSL_CALOCATION"] != "" {
		t.Fatalf("CA location=%q, want empty", env["ZBX_TELEMETRYPROVIDER_0_SSL_CALOCATION"])
	}
	if _, err := os.Stat(filepath.Join(internal, "providers", "telemetry-0", "ca")); !os.IsNotExist(err) {
		t.Fatalf("web telemetry CA was copied into a managed directory: %v", err)
	}
}

func TestResolveSecretsUsesGlobalLocationsForHistoryProviderTLS(t *testing.T) {
	external, internal, _ := useSecretRoots(t)
	providerDir := filepath.Join(external, "providers", "history-0")
	writeTestSecret(t, filepath.Join(providerDir, "ca.pem"), "ca")
	writeTestSecret(t, filepath.Join(providerDir, "client.pem"), "certificate")
	writeTestSecret(t, filepath.Join(providerDir, "client.key"), "key")
	env := Environment{
		"ZABBIX_USER_HOME_DIR":                   t.TempDir(),
		"ZBX_HISTORYPROVIDER_0":                  `{"provider":"clickhouse","url":"https://clickhouse:8443","db":"zabbix"}`,
		"ZBX_HISTORYPROVIDER_0_SSL_CAFILE":       "ca.pem",
		"ZBX_HISTORYPROVIDER_0_SSL_CERTFILE":     "client.pem",
		"ZBX_HISTORYPROVIDER_0_SSL_KEYFILE":      "client.key",
		"ZBX_HISTORYPROVIDER_0_SSL_CALOCATION":   "/user/ca",
		"ZBX_HISTORYPROVIDER_0_SSL_CERTLOCATION": "/user/certs",
		"ZBX_HISTORYPROVIDER_0_SSL_KEYLOCATION":  "/user/keys",
	}
	if err := ResolveSecrets(env, SecretsServerMySQL, SecretsService); err != nil {
		t.Fatal(err)
	}

	locations := map[string]string{
		"ZBX_SSLCALOCATION":   filepath.Join(internal, "ssl", "ca"),
		"ZBX_SSLCERTLOCATION": filepath.Join(internal, "ssl", "certs"),
		"ZBX_SSLKEYLOCATION":  filepath.Join(internal, "ssl", "keys"),
	}
	for variable, want := range locations {
		if env[variable] != want {
			t.Errorf("%s=%q, want %q", variable, env[variable], want)
		}
	}
	for _, variable := range []string{
		"ZBX_HISTORYPROVIDER_0_SSL_CALOCATION",
		"ZBX_HISTORYPROVIDER_0_SSL_CERTLOCATION",
		"ZBX_HISTORYPROVIDER_0_SSL_KEYLOCATION",
	} {
		if _, exists := env[variable]; exists {
			t.Errorf("unsupported provider location %s remains", variable)
		}
	}
	for variable, content := range map[string]string{
		"ZBX_HISTORYPROVIDER_0_SSL_CERTFILE": "certificate",
		"ZBX_HISTORYPROVIDER_0_SSL_KEYFILE":  "key",
	} {
		name := env[variable]
		if filepath.IsAbs(name) {
			t.Errorf("%s=%q, want file name", variable, name)
			continue
		}
		directory := env["ZBX_SSLCERTLOCATION"]
		if strings.HasSuffix(variable, "KEYFILE") {
			directory = env["ZBX_SSLKEYLOCATION"]
		}
		data, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil || string(data) != content {
			t.Errorf("%s content=%q, err=%v", variable, data, err)
		}
	}
	if _, exists := env["ZBX_HISTORYPROVIDER_0_SSL_CAFILE"]; exists {
		t.Error("history provider CA file remains provider-specific")
	}
}

func TestHistoryProviderTLSKeepsVaultFilesInSharedLocations(t *testing.T) {
	external, internal, _ := useSecretRoots(t)
	writeTestSecret(t, filepath.Join(external, "ssl", "certs", "vault.pem"), "vault")
	writeTestSecret(t, filepath.Join(external, "providers", "history-0", "client.pem"), "history")
	env := Environment{
		"ZABBIX_USER_HOME_DIR":               t.TempDir(),
		"ZBX_VAULT":                          "CyberArk",
		"ZBX_VAULTTLSCERTFILE":               "vault.pem",
		"ZBX_HISTORYPROVIDER_0":              `{"provider":"clickhouse","url":"https://clickhouse:8443","db":"zabbix"}`,
		"ZBX_HISTORYPROVIDER_0_SSL_CERTFILE": "client.pem",
	}
	if err := ResolveSecrets(env, SecretsServerMySQL, SecretsService); err != nil {
		t.Fatal(err)
	}
	wantLocation := filepath.Join(internal, "ssl", "certs")
	if env["ZBX_SSLCERTLOCATION"] != wantLocation {
		t.Fatalf("ZBX_SSLCERTLOCATION=%q, want %q", env["ZBX_SSLCERTLOCATION"], wantLocation)
	}
	if env["ZBX_VAULTTLSCERTFILE"] != filepath.Join(wantLocation, "vault.pem") {
		t.Fatalf("ZBX_VAULTTLSCERTFILE=%q", env["ZBX_VAULTTLSCERTFILE"])
	}
	if err := SetVaultTLSFileNames(env); err != nil {
		t.Fatal(err)
	}
	if env["ZBX_VAULTTLSCERTFILE"] != "vault.pem" {
		t.Fatalf("native Vault certificate=%q, want file name", env["ZBX_VAULTTLSCERTFILE"])
	}
}
