package proxy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

// proxyEnvironment prepares the directories and configuration files that the
// proxy images ship, and returns the environment pointing at them.
func proxyEnvironment(t *testing.T) (bootstrap.Environment, string) {
	t.Helper()

	homeDir := t.TempDir()
	configDir := t.TempDir()
	internalCADir := filepath.Join(homeDir, "ssl", "ssl_ca_internal")
	for _, dir := range []string{filepath.Join(homeDir, "ssl", "ssl_ca"), internalCADir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"zabbix_proxy_modules.conf", "zabbix_proxy_telemetry.conf"} {
		if err := os.WriteFile(filepath.Join(configDir, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	return bootstrap.Environment{
		"ZABBIX_USER_HOME_DIR": homeDir,
		"ZABBIX_CONF_DIR":      configDir,
		"ZBX_SSLCALOCATION":    internalCADir,
	}, configDir
}

func TestPrepareServerAndHostname(t *testing.T) {
	tests := []struct {
		name         string
		env          map[string]string
		wantHostname string
		wantServer   string
	}{
		{name: "defaults", wantHostname: "proxy-1", wantServer: "zabbix-server"},
		{
			name:         "explicit hostname and server",
			env:          map[string]string{"ZBX_HOSTNAME": "proxy-a", "ZBX_SERVER_HOST": "zabbix.example"},
			wantHostname: "proxy-a", wantServer: "zabbix.example",
		},
		{
			// The proxy resolves its own name through the item, so the
			// hostname must stay empty.
			name:         "hostname item",
			env:          map[string]string{"ZBX_HOSTNAMEITEM": "system.hostname"},
			wantHostname: "", wantServer: "zabbix-server",
		},
		{
			name:         "hostname wins over the item",
			env:          map[string]string{"ZBX_HOSTNAME": "proxy-b", "ZBX_HOSTNAMEITEM": "system.hostname"},
			wantHostname: "proxy-b", wantServer: "zabbix-server",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env, _ := proxyEnvironment(t)
			for name, value := range test.env {
				env[name] = value
			}

			if err := Prepare(env, "proxy-1"); err != nil {
				t.Fatal(err)
			}
			if env["ZBX_HOSTNAME"] != test.wantHostname {
				t.Errorf("ZBX_HOSTNAME = %q, want %q", env["ZBX_HOSTNAME"], test.wantHostname)
			}
			if env["ZBX_SERVER_HOST"] != test.wantServer {
				t.Errorf("ZBX_SERVER_HOST = %q, want %q", env["ZBX_SERVER_HOST"], test.wantServer)
			}
		})
	}
}

func TestPrepareSNMPTraps(t *testing.T) {
	for _, test := range []struct {
		enable    string
		wantStart string
	}{
		{enable: "", wantStart: ""},
		{enable: "false", wantStart: ""},
		{enable: "true", wantStart: "1"},
		{enable: "True", wantStart: "1"},
	} {
		t.Run("ZBX_ENABLE_SNMP_TRAPS="+test.enable, func(t *testing.T) {
			env, _ := proxyEnvironment(t)
			if test.enable != "" {
				env["ZBX_ENABLE_SNMP_TRAPS"] = test.enable
			}

			if err := Prepare(env, "proxy-1"); err != nil {
				t.Fatal(err)
			}
			if env["ZBX_STARTSNMPTRAPPER"] != test.wantStart {
				t.Errorf("ZBX_STARTSNMPTRAPPER = %q, want %q", env["ZBX_STARTSNMPTRAPPER"], test.wantStart)
			}
			if _, exists := env["ZBX_ENABLE_SNMP_TRAPS"]; exists {
				t.Error("ZBX_ENABLE_SNMP_TRAPS was passed to the proxy")
			}
		})
	}
}

func TestPrepareModulesAndTelemetry(t *testing.T) {
	env, configDir := proxyEnvironment(t)
	env["ZBX_LOADMODULE"] = "dummy.so"
	env["ZBX_TELEMETRYPROVIDER_0"] = `{"provider":"clickhouse","url":"http://clickhouse:8123","db":"zabbix"}`
	env["ZBX_TELEMETRYPROVIDER_0_PASSWORD"] = "secret"
	env["ZBX_TELEMETRYPROVIDER_0_USERNAME"] = "zabbix"

	if err := Prepare(env, "proxy-1"); err != nil {
		t.Fatal(err)
	}

	modules, err := os.ReadFile(filepath.Join(configDir, "zabbix_proxy_modules.conf"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(modules), "LoadModule=dummy.so") {
		t.Errorf("modules configuration = %q", modules)
	}

	telemetry, err := os.ReadFile(filepath.Join(configDir, "zabbix_proxy_telemetry.conf"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(telemetry), "TelemetryProvider=${ZBX_TELEMETRYPROVIDER_0}") {
		t.Errorf("telemetry configuration = %q", telemetry)
	}
	for _, expected := range []string{"clickhouse;", `url="http://clickhouse:8123"`, `password="secret"`} {
		if !strings.Contains(env["ZBX_TELEMETRYPROVIDER_0"], expected) {
			t.Errorf("provider %q does not contain %q", env["ZBX_TELEMETRYPROVIDER_0"], expected)
		}
	}
	for _, name := range []string{"ZBX_TELEMETRYPROVIDER_0_PASSWORD", "ZBX_TELEMETRYPROVIDER_0_USERNAME"} {
		if _, exists := env[name]; exists {
			t.Errorf("%s remains in the environment", name)
		}
	}
}

func TestPrepareRejectsInvalidTelemetryProvider(t *testing.T) {
	env, _ := proxyEnvironment(t)
	env["ZBX_TELEMETRYPROVIDER_1"] = `{"provider":"clickhouse","url":"http://clickhouse:8123","db":"zabbix"}`

	err := Prepare(env, "proxy-1")
	if err == nil || !strings.Contains(err.Error(), "ZBX_TELEMETRYPROVIDER_1 uses index 1, but index 0 is missing") {
		t.Fatalf("error = %v", err)
	}
}

func TestPrepareClearsPrivateEnvironment(t *testing.T) {
	env, _ := proxyEnvironment(t)
	env["DB_SERVER_HOST"] = "mysql"
	env["MYSQL_PASSWORD"] = "secret"
	env["ZBX_DB_PASSWORD"] = "secret"

	if err := Prepare(env, "proxy-1"); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"DB_SERVER_HOST", "MYSQL_PASSWORD", "ZABBIX_USER_HOME_DIR", "ZABBIX_CONF_DIR"} {
		if _, exists := env[name]; exists {
			t.Errorf("%s was passed to the proxy", name)
		}
	}
	// The proxy reads its own database password from the configuration file,
	// which references this variable.
	if env["ZBX_DB_PASSWORD"] != "secret" {
		t.Error("ZBX_DB_PASSWORD must stay in the environment")
	}
}
