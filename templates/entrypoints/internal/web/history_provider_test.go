package web

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

func TestConfigureWebHistoryProviders(t *testing.T) {
	env := bootstrap.Environment{
		"ZBX_HISTORYPROVIDERS":           "[]",
		"ZBX_HISTORYPROVIDER_0":          `{"provider":"clickhouse","types":["uint"],"url":"http://clickhouse:8123","db":"zabbix","username":"inline-user","password":"inline-password","precache":true}`,
		"ZBX_HISTORYPROVIDER_0_USERNAME": "secret-user",
		"ZBX_HISTORYPROVIDER_0_PASSWORD": "p'ass",
		"ZBX_HISTORYPROVIDER_1":          `{"provider":"elasticsearch","types":["str"],"url":"http://elasticsearch:9200","date_index":0}`,
	}
	if err := configureWebHistoryProviders(env); err != nil {
		t.Fatal(err)
	}
	var providers []map[string]any
	if err := json.Unmarshal([]byte(env["ZBX_HISTORYPROVIDERS"]), &providers); err != nil {
		t.Fatal(err)
	}
	if len(providers) != 2 || providers[0]["username"] != "secret-user" || providers[0]["password"] != "p'ass" {
		t.Fatalf("unexpected normalized providers: %#v", providers)
	}
	for _, provider := range providers {
		if _, exists := provider["precache"]; exists {
			t.Fatal("server-only option leaked into the frontend configuration")
		}
		if _, exists := provider["date_index"]; exists {
			t.Fatal("server-only option leaked into the frontend configuration")
		}
	}
	if strings.Contains(env["ZBX_HISTORYPROVIDERS"], "p'ass") {
		t.Fatal("apostrophe was not escaped for the PHP JSON helper")
	}
	for _, name := range []string{"ZBX_HISTORYPROVIDER_0", "ZBX_HISTORYPROVIDER_0_USERNAME", "ZBX_HISTORYPROVIDER_0_PASSWORD", "ZBX_HISTORYPROVIDER_1"} {
		if _, exists := env[name]; exists {
			t.Errorf("%s remains in the environment", name)
		}
	}
	if _, exists := webServerEnv(env)["ZBX_HISTORYPROVIDERS"]; exists {
		t.Fatal("history provider credentials were passed to the web server")
	}
}

func TestConfigureWebHistoryProvidersUnindexedObject(t *testing.T) {
	env := bootstrap.Environment{
		"ZBX_HISTORYPROVIDERS":          `{'provider':'clickhouse','types':['uint'],'url':'http://clickhouse:8123','password':'inline-password'}`,
		"ZBX_HISTORYPROVIDERS_PASSWORD": "secret-password",
	}
	if err := configureWebHistoryProviders(env); err != nil {
		t.Fatal(err)
	}
	var providers []map[string]any
	if err := json.Unmarshal([]byte(env["ZBX_HISTORYPROVIDERS"]), &providers); err != nil {
		t.Fatal(err)
	}
	if len(providers) != 1 || providers[0]["password"] != "secret-password" {
		t.Fatalf("unindexed provider was not normalized: %#v", providers)
	}
}

func TestConfigureWebHistoryProviderSecretFiles(t *testing.T) {
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
	if err := configureWebHistoryProviders(env); err != nil {
		t.Fatal(err)
	}
	var providers []map[string]any
	if err := json.Unmarshal([]byte(env["ZBX_HISTORYPROVIDERS"]), &providers); err != nil {
		t.Fatal(err)
	}
	if len(providers) != 1 || providers[0]["username"] != "secret-user" || providers[0]["password"] != "secret-password" {
		t.Fatalf("Secret file credentials were not used: %#v", providers)
	}
	if _, exists := env["ZBX_HISTORYPROVIDER_0_PASSWORD_FILE"]; exists {
		t.Fatal("Secret file variable remained in the environment")
	}
}

func TestConfigureWebHistoryProvidersRejectsInvalidSettings(t *testing.T) {
	tests := []struct {
		name    string
		env     bootstrap.Environment
		wantErr string
	}{
		{
			name: "mixed formats",
			env: bootstrap.Environment{
				"ZBX_HISTORYPROVIDERS":  `{"provider":"elasticsearch","types":["uint"],"url":"http://elasticsearch:9200"}`,
				"ZBX_HISTORYPROVIDER_0": `{"provider":"clickhouse","types":["str"],"url":"http://clickhouse:8123"}`,
			},
			wantErr: "cannot combine",
		},
		{
			name: "missing index",
			env: bootstrap.Environment{
				"ZBX_HISTORYPROVIDER_1": `{"provider":"elasticsearch","types":["uint"],"url":"http://elasticsearch:9200"}`,
			},
			wantErr: "ZBX_HISTORYPROVIDER_1 uses index 1, but index 0 is missing",
		},
		{
			name: "invalid JSON",
			env: bootstrap.Environment{
				"ZBX_HISTORYPROVIDER_0": "not-json",
			},
			wantErr: "invalid provider JSON",
		},
		{
			name: "credentials without provider",
			env: bootstrap.Environment{
				"ZBX_HISTORYPROVIDERS_PASSWORD": "secret",
			},
			wantErr: "require one unindexed JSON provider",
		},
		{
			name: "credentials for Elasticsearch",
			env: bootstrap.Environment{
				"ZBX_HISTORYPROVIDER_0":          `{"provider":"elasticsearch","types":["uint"],"url":"http://elasticsearch:9200"}`,
				"ZBX_HISTORYPROVIDER_0_PASSWORD": "secret",
			},
			wantErr: "only supported by ClickHouse",
		},
		{
			name: "missing provider type",
			env: bootstrap.Environment{
				"ZBX_HISTORYPROVIDER_0": `{"types":["uint"],"url":"http://clickhouse:8123"}`,
			},
			wantErr: "provider must be clickhouse or elasticsearch",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := configureWebHistoryProviders(test.env)
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("configureWebHistoryProviders() error = %v, want %q", err, test.wantErr)
			}
		})
	}
}
