package storageprovider

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

func TestProviderTLSPaths(t *testing.T) {
	for _, param := range []*Parameter{History, Telemetry} {
		paths := []string{"client.pem", "team/client.pem"}
		if param == Telemetry {
			paths = append(paths, "/custom/team/client.pem")
		}
		for _, path := range paths {
			t.Run(param.Prefix+"/"+path, func(t *testing.T) {
				fields := map[string]any{
					"provider": "clickhouse", "url": "https://clickhouse:8443", "db": "zabbix",
					"ssl_cert_file": path, "ssl_key_file": path,
				}
				raw, err := json.Marshal(fields)
				if err != nil {
					t.Fatal(err)
				}
				name := bootstrap.IndexedName(param.Prefix, 0)
				env := bootstrap.Environment{name: string(raw)}
				configuration, err := NewConfiguration(env, param)
				if err != nil {
					t.Fatal(err)
				}
				if err := configuration.ConfigureNative(); err != nil {
					t.Fatal(err)
				}
				for _, kind := range []string{"cert", "key"} {
					file := path
					location := "ssl_" + kind + "_location="
					if filepath.IsAbs(path) {
						file = filepath.Base(path)
						if !strings.Contains(env[name], location+`"/custom/team"`) {
							t.Errorf("absolute TLS path has no matching directory: %s", env[name])
						}
					} else if strings.Contains(env[name], location) {
						t.Errorf("relative TLS path overrides the standard directory: %s", env[name])
					}
					if !strings.Contains(env[name], "ssl_"+kind+"_file="+quoteOption(file)) {
						t.Errorf("unexpected TLS file in %s", env[name])
					}
				}
			})
		}
		for _, key := range tlsFiles {
			for _, invalid := range []any{"../client.pem", "team/../../client.pem", "", "client\n.pem", nil, 42} {
				fields := map[string]any{"provider": "clickhouse", key: invalid}
				raw, err := json.Marshal(fields)
				if err != nil {
					t.Fatal(err)
				}
				name := bootstrap.IndexedName(param.Prefix, 0)
				configuration, err := NewConfiguration(bootstrap.Environment{name: string(raw)}, param)
				if err != nil {
					t.Fatal(err)
				}
				if err := configuration.ConfigureNative(); err == nil {
					t.Errorf("%s accepted %s=%v", name, key, invalid)
				}
			}
		}
	}
}

func TestValueTypesAreNotRestricted(t *testing.T) {
	for _, value := range []string{
		`{"provider":"elasticsearch","types":["uint",""],"url":"http://elasticsearch:9200"}`,
		"{\"provider\":\"elasticsearch\",\"types\":[\"uint\",\"future\\nLogFile=/tmp/zabbix.log\"],\"url\":\"http://elasticsearch:9200\"}",
	} {
		env := bootstrap.Environment{"ZBX_HISTORYPROVIDER_0": value}
		configuration, err := NewConfiguration(env, History)
		if err != nil {
			t.Fatal(err)
		}
		err = configuration.ConfigureNative()
		if err == nil || !strings.Contains(err.Error(), "must contain value type names") {
			t.Fatalf("error = %v", err)
		}
	}

	env := bootstrap.Environment{
		"ZBX_HISTORYPROVIDER_0": `{"provider":"elasticsearch","types":["uint","future"],"url":"http://elasticsearch:9200"}`,
	}
	configuration, err := NewConfiguration(env, History)
	if err != nil {
		t.Fatal(err)
	}
	if err := configuration.ConfigureNative(); err != nil {
		t.Fatalf("unknown value type was rejected: %v", err)
	}
}

func TestUnsupportedProvidersAreRejected(t *testing.T) {
	for _, test := range []struct {
		param *Parameter
		value string
	}{
		{History, `{"provider":"unknown","types":["uint"],"url":"http://example:9000"}`},
		{Telemetry, `{"provider":"elasticsearch","url":"http://example:9200","db":"zabbix"}`},
	} {
		env := bootstrap.Environment{bootstrap.IndexedName(test.param.Prefix, 0): test.value}
		configuration, err := NewConfiguration(env, test.param)
		if err != nil {
			t.Fatal(err)
		}
		if err := configuration.ConfigureNative(); err == nil {
			t.Errorf("%s accepted an unsupported provider", test.param.Prefix)
		}
	}
}

func TestHistoryInvalidOptions(t *testing.T) {
	for _, test := range []struct {
		field      string
		value      any
		diagnostic string
	}{
		{"vault_path", "secret/history", "only supported by telemetry providers"},
	} {
		t.Run(test.field, func(t *testing.T) {
			fields := map[string]any{"provider": "clickhouse", "url": "https://clickhouse:8443", "db": "zabbix", "types": []string{"uint"}, test.field: test.value}
			raw, err := json.Marshal(fields)
			if err != nil {
				t.Fatal(err)
			}
			configuration, err := NewConfiguration(bootstrap.Environment{
				"ZBX_HISTORYPROVIDER_0":        string(raw),
				"ZBX_HISTORYPROVIDER_0_SCHEMA": `{"create":true}`,
			}, History)
			if err != nil {
				t.Fatal(err)
			}
			_, prepareErr := configuration.ConfigsToPrepare()
			nativeErr := configuration.ConfigureNative()
			_, webErr := WebProviders(bootstrap.Environment{"ZBX_HISTORYPROVIDER_0": string(raw)}, History)
			for _, err := range []error{prepareErr, nativeErr, webErr} {
				if err == nil || !strings.Contains(err.Error(), "ZBX_HISTORYPROVIDER_0") || !strings.Contains(err.Error(), test.field) || !strings.Contains(err.Error(), test.diagnostic) {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestHistoryTLSVariablesSupportAbsoluteFiles(t *testing.T) {
	env := bootstrap.Environment{
		"ZBX_HISTORYPROVIDER_0":                `{"provider":"clickhouse","url":"https://clickhouse:8443","db":"zabbix","types":["uint"]}`,
		"ZBX_HISTORYPROVIDER_0_SSL_CERTFILE":   "/run/zabbix/secrets/providers/history-0/client.pem",
		"ZBX_HISTORYPROVIDER_0_SSL_KEYFILE":    "/run/zabbix/secrets/providers/history-0/client.key",
		"ZBX_HISTORYPROVIDER_0_SSL_CALOCATION": "/run/zabbix/secrets_internal/providers/history-0/ca",
	}
	configuration, err := NewConfiguration(env, History)
	if err != nil {
		t.Fatal(err)
	}
	if err := configuration.ConfigureNative(); err != nil {
		t.Fatal(err)
	}
	for _, option := range []string{
		`ssl_cert_location="/run/zabbix/secrets/providers/history-0"`,
		`ssl_cert_file="client.pem"`,
		`ssl_key_location="/run/zabbix/secrets/providers/history-0"`,
		`ssl_key_file="client.key"`,
		`ssl_ca_location="/run/zabbix/secrets_internal/providers/history-0/ca"`,
	} {
		if !strings.Contains(env["ZBX_HISTORYPROVIDER_0"], option) {
			t.Fatalf("native provider %q does not contain %q", env["ZBX_HISTORYPROVIDER_0"], option)
		}
	}
}

func TestProviderTLSVariableConflictsWithJSONField(t *testing.T) {
	env := bootstrap.Environment{
		"ZBX_HISTORYPROVIDER_0":              `{"provider":"clickhouse","url":"https://clickhouse:8443","db":"zabbix","types":["uint"],"ssl_cert_file":"legacy.pem"}`,
		"ZBX_HISTORYPROVIDER_0_SSL_CERTFILE": "/cert.pem",
	}
	configuration, err := NewConfiguration(env, History)
	if err != nil {
		t.Fatal(err)
	}
	err = configuration.ConfigureNative()
	if err == nil || !strings.Contains(err.Error(), "cannot be combined") {
		t.Fatalf("ConfigureNative() error=%v, want source conflict", err)
	}
}

func TestTelemetrySourceIP(t *testing.T) {
	env := bootstrap.Environment{"ZBX_TELEMETRYPROVIDER_0": `{"provider":"clickhouse","url":"http://clickhouse:8123","db":"zabbix","source_ip":"192.0.2.1"}`}
	configuration, err := NewConfiguration(env, Telemetry)
	if err != nil {
		t.Fatal(err)
	}
	if err := configuration.ConfigureNative(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(env["ZBX_TELEMETRYPROVIDER_0"], `source_ip="192.0.2.1"`) {
		t.Fatal("telemetry source_ip was not preserved")
	}
}

func TestManagedLocations(t *testing.T) {
	for _, test := range []struct {
		param *Parameter
		value string
	}{
		{param: History, value: `{"provider":"clickhouse","types":["uint"],"url":"http://clickhouse:8123","ssl_ca_location":"/ca"}`},
		{param: Telemetry, value: `{"provider":"clickhouse","url":"https://clickhouse:8443","db":"zabbix","ssl_ca_location":"/ca"}`},
	} {
		env := bootstrap.Environment{bootstrap.IndexedName(test.param.Prefix, 0): test.value}
		configuration, err := NewConfiguration(env, test.param)
		if err != nil {
			t.Fatal(err)
		}
		err = configuration.ConfigureNative()
		if err == nil || !strings.Contains(err.Error(), "managed by the image") {
			t.Errorf("%s error = %v", test.param.Prefix, err)
		}
	}
}

func TestHistoryNativeOptions(t *testing.T) {
	for _, test := range []struct{ raw, want string }{
		{`{"provider":"sql","types":["dbl","uint"],"precache":0}`, `sql;precache=0,value_types="dbl,uint"`},
		{`{"provider":"clickhouse","source_ip":"192.0.2.1","log_slow_queries":3000}`, `clickhouse;log_slow_queries=3000,source_ip="192.0.2.1"`},
		{`{"provider":"elasticsearch","source_ip":"192.0.2.1","log_slow_queries":3000}`, `elasticsearch;log_slow_queries=3000,source_ip="192.0.2.1"`},
	} {
		env := bootstrap.Environment{"ZBX_HISTORYPROVIDER_0": test.raw}
		configuration, err := NewConfiguration(env, History)
		if err != nil {
			t.Fatal(err)
		}
		if err := configuration.ConfigureNative(); err != nil {
			t.Fatal(err)
		}
		if got := env["ZBX_HISTORYPROVIDER_0"]; got != test.want {
			t.Errorf("got %q, want %q", got, test.want)
		}
	}
}

func TestWebSQLHistory(t *testing.T) {
	for _, env := range []bootstrap.Environment{
		{"ZBX_HISTORYPROVIDER_0": `{"provider":"sql","types":["dbl","uint"],"precache":0}`},
		{"ZBX_HISTORYPROVIDERS": `[{"provider":"sql","types":["dbl","uint"],"precache":0}]`},
	} {
		providers, err := WebProviders(env, History)
		if err != nil {
			t.Fatal(err)
		}
		if len(providers) != 0 {
			t.Fatalf("SQL provider passed to frontend: %v", providers)
		}
	}
}
