package storageinit

import (
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/testutil"
)

func TestProviderCredentials(t *testing.T) {
	queries := make(map[string][]string)
	newEndpoint := func(label string) string {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			queries[label] = append(queries[label], string(body))
			user, pass, _ := r.BasicAuth()
			if user != label || pass != "file-secret" {
				t.Errorf("incorrect credentials for %s", label)
			}
			fmt.Fprintln(w, "1")
		}))
		t.Cleanup(server.Close)
		return server.URL
	}
	secret := filepath.Join(t.TempDir(), "password")
	if err := os.WriteFile(secret, []byte("file-secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	env := bootstrap.Environment{}
	for _, spec := range []struct{ name, db, types string }{
		{"ZBX_HISTORYPROVIDER_0", "numbers", `,"types":["uint","dbl"]`},
		{"ZBX_HISTORYPROVIDER_1", "logs", `,"types":["log"]`},
		{"ZBX_TELEMETRYPROVIDER_0", "telemetry", ""},
	} {
		env[spec.name] = fmt.Sprintf(`{"provider":"clickhouse","url":%q,"db":%q%s}`, newEndpoint(spec.db), spec.db, spec.types)
		env[spec.name+"_USERNAME"] = spec.db
		env[spec.name+"_PASSWORD_FILE"] = secret
		env[spec.name+"_SCHEMA"] = `{"create":false}`
	}
	logs := testutil.CaptureStdout(t, func() {
		if err := ConfigureHistory(env); err != nil {
			t.Fatal(err)
		}
		if err := ConfigureTelemetry(env); err != nil {
			t.Fatal(err)
		}
	})
	for label, count := range map[string]int{"numbers": 3, "logs": 2, "telemetry": 8} {
		if len(queries[label]) != count {
			t.Fatalf("%s requests: %v", label, queries[label])
		}
		for _, q := range queries[label] {
			if !strings.Contains(q, "'"+label+"'") {
				t.Errorf("wrong database for %s: %s", label, q)
			}
			if strings.HasPrefix(q, "CREATE") {
				t.Fatalf("check-only request wrote DDL: %s", q)
			}
		}
	}
	for _, name := range []string{"ZBX_HISTORYPROVIDER_0", "ZBX_HISTORYPROVIDER_1", "ZBX_TELEMETRYPROVIDER_0"} {
		for _, suffix := range []string{"_USERNAME", "_PASSWORD"} {
			message := "Using " + name + suffix + " variable"
			if count := strings.Count(logs, message); count != 1 {
				t.Errorf("%s logged %d times, want once", name+suffix, count)
			}
		}
		if !strings.Contains(env[name], `password="file-secret"`) {
			t.Errorf("%s does not contain the resolved password", name)
		}
	}
	if strings.Contains(logs, "file-secret") {
		t.Error("logs contain the password")
	}

	for name, value := range env {
		if strings.HasSuffix(name, "_SCHEMA") || strings.HasSuffix(name, "_PASSWORD_FILE") {
			t.Errorf("variable was not removed: %s", name)
		}
		if strings.Contains(value, "create=") || strings.Contains(value, "cluster=") {
			t.Error("native configuration contains _SCHEMA options")
		}
	}
}

func TestSchemaOptions(t *testing.T) {
	base := bootstrap.Environment{"ZBX_HISTORYPROVIDER_0": `{"provider":"clickhouse","url":"http://unused:8123","db":"data","types":["uint"]}`}
	for _, raw := range []string{`{}`, `null`, `{"create":"true"}`, `{"create":true,"clustr":"x"}`, `{"create":true,"ttl":-1}`, `{"create":true,"partition":"unsupported"}`, `{"create":false} {}`} {
		t.Run(raw, func(t *testing.T) {
			env := maps.Clone(base)
			env["ZBX_HISTORYPROVIDER_0_SCHEMA"] = raw
			prepareCADirs(t, env)
			if _, err := PrepareHistory(env); err == nil {
				t.Fatal("invalid options accepted")
			}
		})
	}
	for name, value := range map[string]string{
		"native":           `clickhouse;url=http://unused:8123,db=data,value_types="uint"`,
		"other provider":   `{"provider":"unknown","url":"http://unused:9200","types":["uint"]}`,
		"missing database": `{"provider":"clickhouse","url":"http://unused:8123","types":["uint"]}`,
	} {
		t.Run(name, func(t *testing.T) {
			env := bootstrap.Environment{"ZBX_HISTORYPROVIDER_0": value, "ZBX_HISTORYPROVIDER_0_SCHEMA": `{"create":true}`}
			prepareCADirs(t, env)
			if _, err := PrepareHistory(env); err == nil {
				t.Fatal("invalid provider accepted")
			}
		})
	}
}

func TestNoSchemaNoConnection(t *testing.T) {
	env := bootstrap.Environment{"ZBX_TELEMETRYPROVIDER_0": `{"provider":"clickhouse","url":"http://unused:8123","db":"data"}`}
	if count, err := PrepareTelemetry(env); err != nil || count != 0 {
		t.Fatalf("prepared %d providers without _SCHEMA: %v", count, err)
	}
}

func TestSchemaVaultCredentials(t *testing.T) {
	chCalls, vaultCalls := 0, 0
	clickhouse := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		chCalls++
		username, password, ok := r.BasicAuth()
		if !ok || username != "vault-user" || password != "vault-secret" {
			t.Error("Vault credentials not used")
		}
		fmt.Fprintln(w, "1")
	}))
	defer clickhouse.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vaultCalls++
		if r.URL.Path != "/v1/secret/data/telemetry" || r.Header.Get("X-Vault-Token") != "test-token" {
			t.Errorf("unexpected Vault request: %s", r.URL.Path)
		}
		fmt.Fprint(w, `{"data":{"data":{"username":"vault-user","password":"vault-secret"}}}`)
	}))
	defer server.Close()
	env := bootstrap.Environment{
		"ZBX_VAULT": "HashiCorp", "ZBX_VAULTURL": server.URL, "VAULT_TOKEN": "test-token",
		"ZBX_TELEMETRYPROVIDER_0": fmt.Sprintf(`{"provider":"clickhouse","url":%q,"db":"telemetry","vault_path":"secret/telemetry"}`, clickhouse.URL),
	}
	env["ZBX_TELEMETRYPROVIDER_0_SCHEMA"] = `{"create":false}`
	if err := ConfigureTelemetry(env); err != nil {
		t.Fatal(err)
	}
	if vaultCalls != 1 || chCalls == 0 {
		t.Fatalf("Vault requests=%d, ClickHouse requests=%d", vaultCalls, chCalls)
	}
	if strings.Contains(env["ZBX_TELEMETRYPROVIDER_0"], "vault-secret") || env["ZBX_VAULTDBPATH"] != "" {
		t.Fatal("schema initialization mutated the service's Vault configuration")
	}

	env["ZBX_HISTORYPROVIDER_0"] = fmt.Sprintf(`{"provider":"clickhouse","url":%q,"db":"history","types":["uint"],"username":"vault-user","password":"vault-secret","vault_path":"secret/history"}`, clickhouse.URL)
	env["ZBX_HISTORYPROVIDER_0_SCHEMA"] = `{"create":false}`
	chCalls = 0
	if err := ConfigureHistory(env); err != nil {
		t.Fatal(err)
	}
	if vaultCalls != 1 || chCalls == 0 {
		t.Fatalf("history preparation: Vault requests=%d, ClickHouse requests=%d", vaultCalls-1, chCalls)
	}
	if !strings.Contains(env["ZBX_HISTORYPROVIDER_0"], `vault_path="secret/history"`) {
		t.Fatal("history vault_path was removed from the server configuration")
	}
}

func TestInvalidConnectionFields(t *testing.T) {
	for _, field := range []string{"url", "db", "username", "password", "source_ip", "ssl_key_password", "vault_path"} {
		t.Run(field, func(t *testing.T) {
			fields := map[string]any{"provider": "clickhouse", "url": "http://unused:8123", "db": "data"}
			fields[field] = map[string]any{}
			raw, err := json.Marshal(fields)
			if err != nil {
				t.Fatal(err)
			}
			env := bootstrap.Environment{
				"ZBX_TELEMETRYPROVIDER_0":        string(raw),
				"ZBX_TELEMETRYPROVIDER_0_SCHEMA": `{"create":true}`,
			}
			prepareCADirs(t, env)
			_, err = PrepareTelemetry(env)
			if err == nil || !strings.Contains(err.Error(), field+" must") ||
				!strings.Contains(err.Error(), "ZBX_TELEMETRYPROVIDER_0") {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestMixedProviderFormats(t *testing.T) {
	native := `clickhouse;url=http://unused:8123,db=data,value_types="uint"`
	env := bootstrap.Environment{
		"ZBX_HISTORYPROVIDER_0":          native,
		"ZBX_HISTORYPROVIDER_1":          `{"provider":"clickhouse","url":"http://unused:8123","db":"data","types":["log"]}`,
		"ZBX_HISTORYPROVIDER_1_PASSWORD": "secret",
	}
	if err := ConfigureHistory(env); err != nil {
		t.Fatal(err)
	}
	if env["ZBX_HISTORYPROVIDER_0"] != native {
		t.Error("native configuration changed")
	}
	if !strings.HasPrefix(env["ZBX_HISTORYPROVIDER_1"], "clickhouse;") || !strings.Contains(env["ZBX_HISTORYPROVIDER_1"], `password="secret"`) {
		t.Error("JSON provider was not converted")
	}
	if _, exists := env["ZBX_HISTORYPROVIDER_1_PASSWORD"]; exists {
		t.Error("password variable was not removed")
	}
}

func TestInitSkipsProviderWithoutSchema(t *testing.T) {
	env := bootstrap.Environment{
		"ZBX_HISTORYPROVIDER_0":               `{"provider":"unknown"}`,
		"ZBX_HISTORYPROVIDER_0_PASSWORD_FILE": "/missing/secret",
	}
	if count, err := PrepareHistory(env); err != nil || count != 0 {
		t.Fatalf("prepared %d providers without _SCHEMA: %v", count, err)
	}
	if env["ZBX_HISTORYPROVIDER_0_PASSWORD_FILE"] != "/missing/secret" {
		t.Error("init changed credentials for a provider without _SCHEMA")
	}
}

func TestValidateAllProvidersBeforeRequests(t *testing.T) {
	for _, test := range []struct{ name, provider, options string }{
		{"ClickHouse options", `{"provider":"clickhouse","url":"http://unused:8123","db":"history","types":["uint"]}`, `{}`},
		{"Elasticsearch options", `{"provider":"elasticsearch","url":"http://unused:9200","types":["uint"]}`, `{}`},
		{"ClickHouse URL", `{"provider":"clickhouse","url":"invalid","db":"history","types":["uint"]}`, `{"create":true}`},
		{"Elasticsearch URL", `{"provider":"elasticsearch","url":"invalid","types":["uint"]}`, `{"create":true}`},
		{"ClickHouse CA directory", `{"provider":"clickhouse","url":"https://unused:8123","db":"history","types":["uint"]}`, `{"create":true}`},
		{"ClickHouse missing types", `{"provider":"clickhouse","url":"http://unused:8123","db":"history"}`, `{"create":true}`},
		{"Elasticsearch missing types", `{"provider":"elasticsearch","url":"http://unused:9200"}`, `{"create":true}`},
		{"SQL schema", `{"provider":"sql","types":["uint"]}`, `{"create":true}`},
		{"ClickHouse history type", `{"provider":"clickhouse","url":"http://unused:8123","db":"history","types":["unknown"]}`, `{"create":true}`},
		{"Elasticsearch history type", `{"provider":"elasticsearch","url":"http://unused:9200","types":["json_pipeline"]}`, `{"create":true}`},
		{"history certificate", `{"provider":"clickhouse","url":"http://unused:8123","db":"history","types":["uint"],"ssl_cert_file":"/custom/client.pem"}`, `{"create":true}`},
		{"history key", `{"provider":"clickhouse","url":"http://unused:8123","db":"history","types":["uint"],"ssl_key_file":"/custom/client.key"}`, `{"create":true}`},
		{"history without schema", `{"provider":"elasticsearch","url":"http://unused:9200","types":["uint"],"ssl_ca_location":"/custom/ca"}`, ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			endpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				fmt.Fprintln(w, "1")
			}))
			defer endpoint.Close()
			env := bootstrap.Environment{
				"ZBX_HISTORYPROVIDER_0":        fmt.Sprintf(`{"provider":"clickhouse","url":%q,"db":"history","types":["uint"]}`, endpoint.URL),
				"ZBX_HISTORYPROVIDER_0_SCHEMA": `{"create":true}`,
				"ZBX_HISTORYPROVIDER_1":        test.provider,
				"ZBX_HISTORYPROVIDER_1_SCHEMA": test.options,
				"ZBX_SSLCALOCATION":            filepath.Join(t.TempDir(), "missing"),
			}
			if err := ConfigureHistory(env); err == nil || !strings.Contains(err.Error(), "ZBX_HISTORYPROVIDER_1") {
				t.Fatalf("expected second provider error, got %v", err)
			}
			if calls != 0 {
				t.Fatal("database contacted before all providers were validated")
			}
		})
	}
}
func prepareCADirs(t *testing.T, env bootstrap.Environment) {
	t.Helper()
	home := t.TempDir()
	env["ZABBIX_USER_HOME_DIR"] = home
	env["ZBX_SSLCALOCATION"] = filepath.Join(home, "ssl", "ssl_ca_internal")
	for _, dir := range []string{filepath.Join(home, "ssl", "ssl_ca"), env["ZBX_SSLCALOCATION"]} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal(err)
		}
	}
}
