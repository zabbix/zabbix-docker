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
		for _, path := range []string{"client.pem", "team/client.pem", "/custom/team/client.pem"} {
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

func TestNativeConnectionOptions(t *testing.T) {
	value := `{"provider":"clickhouse","types":["uint"],"url":"http://clickhouse:8123","db":"zabbix","source_ip":"192.0.2.1","log_slow_queries":3000}`
	env := bootstrap.Environment{"ZBX_HISTORYPROVIDER_0": value}
	configuration, err := NewConfiguration(env, History)
	if err != nil {
		t.Fatal(err)
	}
	if err := configuration.ConfigureNative(); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{`source_ip="192.0.2.1"`, "log_slow_queries=3000"} {
		if !strings.Contains(env["ZBX_HISTORYPROVIDER_0"], expected) {
			t.Errorf("provider %q does not contain %q", env["ZBX_HISTORYPROVIDER_0"], expected)
		}
	}

	web := bootstrap.Environment{"ZBX_HISTORYPROVIDER_0": value}
	providers, err := WebProviders(web, History)
	if err != nil {
		t.Fatal(err)
	}
	for _, serverOnly := range []string{"source_ip", "log_slow_queries"} {
		if _, exists := providers[0][serverOnly]; exists {
			t.Errorf("%s reached the frontend", serverOnly)
		}
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
