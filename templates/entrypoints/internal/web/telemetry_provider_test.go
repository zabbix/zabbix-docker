package web

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

func TestWebTelemetryTLSPaths(t *testing.T) {
	homeDir := t.TempDir()
	for _, path := range []string{"team/client.pem", "/custom/team/client.pem"} {
		t.Run(path, func(t *testing.T) {
			raw, err := json.Marshal(map[string]any{
				"provider": "clickhouse", "ssl_cert_file": path, "ssl_key_file": path,
			})
			if err != nil {
				t.Fatal(err)
			}
			env := bootstrap.Environment{"ZBX_TELEMETRYPROVIDER_0": string(raw)}
			if err := configureWebTelemetryProviders(env, homeDir); err != nil {
				t.Fatal(err)
			}
			var providers []map[string]any
			if err := json.Unmarshal([]byte(env["ZBX_TELEMETRYPROVIDERS"]), &providers); err != nil {
				t.Fatal(err)
			}
			for field, directory := range map[string]string{"ssl_cert_file": "certs", "ssl_key_file": "keys"} {
				want := path
				if !filepath.IsAbs(path) {
					want = filepath.Join(homeDir, "ssl", directory, path)
				}
				if got := providers[0][field]; got != want {
					t.Errorf("%s = %v, want %s", field, got, want)
				}
			}
		})
	}
}

func TestConfigureWebTelemetryProviders(t *testing.T) {
	homeDir := t.TempDir()
	env := bootstrap.Environment{
		"ZBX_TELEMETRYPROVIDER_0":                  `{"provider":"clickhouse","url":"https://clickhouse:8443","db":"zabbix","source_ip":"192.0.2.1","ssl_cert_file":"client.crt","ssl_key_file":"client.key"}`,
		"ZBX_TELEMETRYPROVIDER_0_USERNAME":         "zabbix",
		"ZBX_TELEMETRYPROVIDER_0_PASSWORD":         "secret",
		"ZBX_TELEMETRYPROVIDER_0_SSL_KEY_PASSWORD": "key-secret",
	}
	if err := configureWebTelemetryProviders(env, homeDir); err != nil {
		t.Fatal(err)
	}

	var providers []map[string]any
	if err := json.Unmarshal([]byte(env["ZBX_TELEMETRYPROVIDERS"]), &providers); err != nil {
		t.Fatal(err)
	}
	if len(providers) != 1 {
		t.Fatalf("providers = %#v", providers)
	}
	got := providers[0]
	want := map[string]any{
		"ssl_cert_file":    filepath.Join(homeDir, "ssl", "certs", "client.crt"),
		"ssl_key_file":     filepath.Join(homeDir, "ssl", "keys", "client.key"),
		"ssl_ca_location":  "",
		"ssl_key_password": "key-secret",
		"ssl_verify_peer":  true,
		"ssl_verify_host":  true,
		"username":         "zabbix",
		"password":         "secret",
	}
	for field, expected := range want {
		if got[field] != expected {
			t.Errorf("%s = %#v, want %#v", field, got[field], expected)
		}
	}
	if _, exists := got["source_ip"]; exists {
		t.Fatal("server-only source_ip was passed to the frontend")
	}
	for name := range env {
		if strings.HasPrefix(name, "ZBX_TELEMETRYPROVIDER_") {
			t.Errorf("%s remains in the environment", name)
		}
	}
}

// The frontend reads every option of a telemetry data source without a default
// of its own, so a minimal configuration must still produce the full set.
func TestWebTelemetryProvidesEveryOptionTheFrontendReads(t *testing.T) {
	env := bootstrap.Environment{
		"ZBX_TELEMETRYPROVIDER_0": `{"provider":"clickhouse","url":"http://clickhouse:8123","db":"zabbix"}`,
	}
	if err := configureWebTelemetryProviders(env, t.TempDir()); err != nil {
		t.Fatal(err)
	}
	var providers []map[string]any
	if err := json.Unmarshal([]byte(env["ZBX_TELEMETRYPROVIDERS"]), &providers); err != nil {
		t.Fatal(err)
	}
	if len(providers) != 1 {
		t.Fatalf("providers = %#v", providers)
	}
	want := map[string]any{
		"provider": "clickhouse", "url": "http://clickhouse:8123", "db": "zabbix",
		"username": "", "password": "", "vault_path": "",
		"ssl_cert_file": "", "ssl_key_file": "", "ssl_key_password": "",
		"ssl_ca_file": "", "ssl_ca_location": "",
		"ssl_verify_peer": false, "ssl_verify_host": false,
	}
	for option, expected := range want {
		value, exists := providers[0][option]
		if !exists {
			t.Errorf("%s is missing; the frontend reads it without a default", option)
			continue
		}
		if value != expected {
			t.Errorf("%s = %#v, want %#v", option, value, expected)
		}
	}
	if len(providers[0]) != len(want) {
		t.Errorf("unexpected options: %#v", providers[0])
	}
}
