//go:build !windows

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

func TestPrepareService(t *testing.T) {
	env, runtimeRulePath := serviceFixture(t)
	env["ZBX_TLSCERT"] = "certificate-data"
	env["ZBX_TLSKEY"] = "key-data"
	env["MYSQL_PASSWORD"] = "password"
	if err := prepareService(env, runtimeRulePath); err != nil {
		t.Fatal(err)
	}

	if env["ZBX_ALLOWEDIP"] != "127.0.0.1,::1" {
		t.Fatalf("unexpected allowed IP: %q", env["ZBX_ALLOWEDIP"])
	}
	if _, found := env["MYSQL_PASSWORD"]; found {
		t.Fatal("MYSQL_PASSWORD was not removed")
	}
	for _, variable := range []string{"ZBX_TLSCERT", "ZBX_TLSKEY"} {
		if _, found := env[variable]; found {
			t.Fatalf("%s was not removed", variable)
		}
	}
	for variable, expected := range map[string]string{
		"ZBX_TLSCERTFILE": "certificate-data",
		"ZBX_TLSKEYFILE":  "key-data",
	} {
		path := env[variable]
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if string(data) != expected {
			t.Fatalf("unexpected %s content: %q", variable, data)
		}
	}
}

func TestPrepareServiceKeepsExplicitAllowedIP(t *testing.T) {
	env, runtimeRulePath := serviceFixture(t)
	homeDir := env["ZABBIX_USER_HOME_DIR"]
	env["ZBX_ALLOWEDIP"] = "192.0.2.1"
	env["ZBX_CLEAR_ENV"] = "false"

	if err := prepareService(env, runtimeRulePath); err != nil {
		t.Fatal(err)
	}
	if env["ZBX_ALLOWEDIP"] != "192.0.2.1" {
		t.Fatalf("unexpected allowed IP: %q", env["ZBX_ALLOWEDIP"])
	}
	if env["ZABBIX_USER_HOME_DIR"] != homeDir {
		t.Fatal("private environment was cleared despite ZBX_CLEAR_ENV=false")
	}
}

func serviceFixture(t *testing.T) (bootstrap.Environment, string) {
	t.Helper()
	homeDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(homeDir, "enc_internal"), 0o700); err != nil {
		t.Fatal(err)
	}
	configDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(configDir, "zabbix_mcp_server_tool_access.conf"),
		[]byte("DenyTool=item_get\n"), 0o444,
	); err != nil {
		t.Fatal(err)
	}
	return bootstrap.Environment{
		"ZABBIX_USER_HOME_DIR": homeDir,
		"ZABBIX_CONF_DIR":      configDir,
	}, filepath.Join(t.TempDir(), "zabbix_mcp_server_tool_access.conf")
}
