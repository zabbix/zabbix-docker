package server

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

func TestConfigureJSONHistoryProviders(t *testing.T) {
	env := bootstrap.Environment{
		"ZBX_HISTORYPROVIDER_0":          `{"provider":"clickhouse","types":["uint","dbl"],"url":"http://clickhouse:8123","db":"zabbix","username":"inline-user","password":"inline-password","precache":true}`,
		"ZBX_HISTORYPROVIDER_0_USERNAME": "secret-user",
		"ZBX_HISTORYPROVIDER_0_PASSWORD": `secret,"password`,
		"ZBX_HISTORYPROVIDER_1":          `{"provider":"elasticsearch","types":["str","log"],"url":"http://elasticsearch:9200","date_index":0}`,
	}
	if err := configureJSONHistoryProviders(env); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"ZBX_HISTORYPROVIDER_0": `clickhouse;value_types="uint,dbl",url="http://clickhouse:8123",db="zabbix",password="secret,\"password",precache=1,username="secret-user"`,
		"ZBX_HISTORYPROVIDER_1": `elasticsearch;value_types="str,log",url="http://elasticsearch:9200",date_index=0`,
	}
	for name, expected := range want {
		if env[name] != expected {
			t.Errorf("%s = %q, want %q", name, env[name], expected)
		}
	}
	for _, name := range []string{"ZBX_HISTORYPROVIDER_0_USERNAME", "ZBX_HISTORYPROVIDER_0_PASSWORD"} {
		if _, exists := env[name]; exists {
			t.Errorf("%s remains in the environment", name)
		}
	}
}

func TestPrepareWritesJSONHistoryProviderReferences(t *testing.T) {
	homeDir := t.TempDir()
	configDir := t.TempDir()
	for _, dir := range []string{filepath.Join(homeDir, "ssl", "ssl_ca"), filepath.Join(homeDir, "ssl", "ssl_ca_internal")} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, filename := range []string{"zabbix_server_modules.conf", "zabbix_server_history_storage.conf"} {
		if err := os.WriteFile(filepath.Join(configDir, filename), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	env := bootstrap.Environment{
		"ZABBIX_USER_HOME_DIR":  homeDir,
		"ZABBIX_CONF_DIR":       configDir,
		"ZBX_SSLCALOCATION":     filepath.Join(homeDir, "ssl", "ssl_ca_internal"),
		"ZBX_HISTORYPROVIDER_0": `{"provider":"elasticsearch","types":["uint"],"url":"http://elasticsearch:9200"}`,
	}
	if err := Prepare(env); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(configDir, "zabbix_server_history_storage.conf"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "HistoryProvider=${ZBX_HISTORYPROVIDER_0}") {
		t.Fatalf("HistoryProvider reference missing from configuration: %q", data)
	}
	if !strings.HasPrefix(env["ZBX_HISTORYPROVIDER_0"], `elasticsearch;value_types="uint"`) {
		t.Fatalf("converted provider missing from environment: %q", env["ZBX_HISTORYPROVIDER_0"])
	}
}

func TestJSONHistoryProviderInlineCredentialsFallback(t *testing.T) {
	env := bootstrap.Environment{
		"ZBX_HISTORYPROVIDER_0": `{"provider":"clickhouse","types":["uint"],"url":"http://clickhouse:8123","username":"inline-user","password":"inline-password"}`,
	}
	if err := configureJSONHistoryProviders(env); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{`username="inline-user"`, `password="inline-password"`} {
		if !strings.Contains(env["ZBX_HISTORYPROVIDER_0"], expected) {
			t.Fatalf("inline credentials were not preserved: %q", env["ZBX_HISTORYPROVIDER_0"])
		}
	}
}

func TestJSONHistoryProviderSecretFiles(t *testing.T) {
	secretDir := t.TempDir()
	usernameFile := filepath.Join(secretDir, "username")
	passwordFile := filepath.Join(secretDir, "password")
	for path, value := range map[string]string{usernameFile: "secret-user\n", passwordFile: "secret-password\n"} {
		if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	env := bootstrap.Environment{
		"ZBX_HISTORYPROVIDER_0":               `{"provider":"clickhouse","types":["uint"],"url":"http://clickhouse:8123"}`,
		"ZBX_HISTORYPROVIDER_0_USERNAME_FILE": usernameFile,
		"ZBX_HISTORYPROVIDER_0_PASSWORD_FILE": passwordFile,
	}
	if err := configureJSONHistoryProviders(env); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{`username="secret-user"`, `password="secret-password"`} {
		if !strings.Contains(env["ZBX_HISTORYPROVIDER_0"], expected) {
			t.Fatalf("Secret file credentials were not used: %q", env["ZBX_HISTORYPROVIDER_0"])
		}
	}
	if _, exists := env["ZBX_HISTORYPROVIDER_0_PASSWORD_FILE"]; exists {
		t.Fatal("Secret file variable remained in the environment")
	}
}

func TestJSONHistoryProviderTLS(t *testing.T) {
	env := bootstrap.Environment{
		"ZBX_HISTORYPROVIDER_0":                  `{"provider":"clickhouse","types":["uint"],"url":"https://clickhouse:8443","db":"zabbix","ssl_cert_file":"client.crt","ssl_key_file":"client.key","ssl_verify_peer":true,"ssl_verify_host":true}`,
		"ZBX_HISTORYPROVIDER_0_SSL_KEY_PASSWORD": "key-secret",
	}
	if err := configureJSONHistoryProviders(env); err != nil {
		t.Fatal(err)
	}
	value := env["ZBX_HISTORYPROVIDER_0"]
	for _, expected := range []string{
		`ssl_cert_file="client.crt"`, `ssl_key_file="client.key"`,
		`ssl_key_password="key-secret"`, "ssl_verify_peer=1", "ssl_verify_host=1",
	} {
		if !strings.Contains(value, expected) {
			t.Errorf("provider %q does not contain %q", value, expected)
		}
	}
	if strings.Contains(value, "ssl_cert_location") {
		t.Fatalf("provider contains a managed location: %q", value)
	}
}

func TestJSONHistoryProviderAllowsOneCredential(t *testing.T) {
	env := bootstrap.Environment{
		"ZBX_HISTORYPROVIDER_0":          `{"provider":"clickhouse","types":["uint"],"url":"http://clickhouse:8123"}`,
		"ZBX_HISTORYPROVIDER_0_PASSWORD": "secret-password",
	}
	if err := configureJSONHistoryProviders(env); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(env["ZBX_HISTORYPROVIDER_0"], `password="secret-password"`) {
		t.Fatalf("password was not used: %q", env["ZBX_HISTORYPROVIDER_0"])
	}
}

func TestNativeElasticsearchHistoryProviderRemainsSupported(t *testing.T) {
	want := `elasticsearch;value_types="uint",url=http://elasticsearch:9200`
	env := bootstrap.Environment{"ZBX_HISTORYPROVIDER_0": want}
	if err := configureJSONHistoryProviders(env); err != nil {
		t.Fatal(err)
	}
	if env["ZBX_HISTORYPROVIDER_0"] != want {
		t.Fatalf("native Elasticsearch provider changed: %q", env["ZBX_HISTORYPROVIDER_0"])
	}
}

func TestJSONAndNativeHistoryProvidersCanBeCombined(t *testing.T) {
	native := `elasticsearch;value_types="str",url=http://elasticsearch:9200`
	env := bootstrap.Environment{
		"ZBX_HISTORYPROVIDER_0": `{"provider":"elasticsearch","types":["uint"],"url":"http://elasticsearch:9200"}`,
		"ZBX_HISTORYPROVIDER_1": native,
	}
	if err := configureJSONHistoryProviders(env); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(env["ZBX_HISTORYPROVIDER_0"], `elasticsearch;value_types="uint"`) {
		t.Fatalf("JSON provider was not converted: %q", env["ZBX_HISTORYPROVIDER_0"])
	}
	if env["ZBX_HISTORYPROVIDER_1"] != native {
		t.Fatalf("native provider changed: %q", env["ZBX_HISTORYPROVIDER_1"])
	}
}

func TestJSONHistoryProviderValidation(t *testing.T) {
	tests := []struct {
		name    string
		env     bootstrap.Environment
		wantErr string
	}{
		{
			name: "missing index",
			env: bootstrap.Environment{
				"ZBX_HISTORYPROVIDER_1": `{"provider":"elasticsearch","types":["uint"],"url":"http://elasticsearch:9200"}`,
			},
			wantErr: "ZBX_HISTORYPROVIDER_1 uses index 1, but index 0 is missing",
		},
		{
			name: "orphan credentials",
			env: bootstrap.Environment{
				"ZBX_HISTORYPROVIDER_0_PASSWORD": "secret",
			},
			wantErr: "matching JSON provider",
		},
		{
			name: "direct value and Secret file",
			env: bootstrap.Environment{
				"ZBX_HISTORYPROVIDER_0":               `{"provider":"clickhouse","types":["uint"],"url":"http://clickhouse:8123"}`,
				"ZBX_HISTORYPROVIDER_0_USERNAME":      "direct-user",
				"ZBX_HISTORYPROVIDER_0_USERNAME_FILE": "/run/secrets/username",
				"ZBX_HISTORYPROVIDER_0_PASSWORD":      "direct-password",
			},
			wantErr: "both variables",
		},
		{
			name: "credentials with native value",
			env: bootstrap.Environment{
				"ZBX_HISTORYPROVIDER_0":               `clickhouse;value_types="uint",url=http://clickhouse:8123`,
				"ZBX_HISTORYPROVIDER_0_PASSWORD_FILE": "/run/secrets/password",
			},
			wantErr: "credentials require a JSON provider",
		},
		{
			name: "invalid option name",
			env: bootstrap.Environment{
				"ZBX_HISTORYPROVIDER_0": `{"provider":"elasticsearch","types":["uint"],"url":"http://elasticsearch:9200","bad,key":"value"}`,
			},
			wantErr: "invalid option",
		},
		{
			name: "duplicate value types",
			env: bootstrap.Environment{
				"ZBX_HISTORYPROVIDER_0": `{"provider":"elasticsearch","types":["uint"],"value_types":"dbl","url":"http://elasticsearch:9200"}`,
			},
			wantErr: "use types instead of value_types",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := configureJSONHistoryProviders(test.env)
			if test.wantErr == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("configureJSONHistoryProviders() error = %v, want %q", err, test.wantErr)
			}
		})
	}
}
