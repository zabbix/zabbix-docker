package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

func TestPrepareServiceWindows(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	root := t.TempDir()
	configDir := filepath.Join(root, "conf")
	homeDir := filepath.Join(root, "home")
	pluginDirectory := filepath.Join(configDir, "zabbix_agent2.d", "plugins.d")
	if err := os.MkdirAll(pluginDirectory, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(homeDir, "enc_internal"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "zabbix_agent2_item_keys.conf"), []byte("# DenyKey=system.run[*]\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"mongodb.conf", "postgresql.conf", "mssql.conf", "ember.conf"} {
		if err := os.WriteFile(filepath.Join(pluginDirectory, name), []byte("# plugin config\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	env := bootstrap.Environment{
		"ZABBIX_CONF_DIR": configDir, "ZABBIX_USER_HOME_DIR": homeDir,
		"ZBX_ENABLESTATUSPORT": "true", "ZBX_STATUSPORT": "12345",
		"UNRELATED_VARIABLE": "value",
	}
	if err := prepareService(env); err != nil {
		t.Fatal(err)
	}
	if env["ZBX_STATUSPORT"] != "12345" {
		t.Fatalf("unexpected Windows status port: %q", env["ZBX_STATUSPORT"])
	}
	if env["UNRELATED_VARIABLE"] != "value" {
		t.Fatal("Windows entrypoint removed an unrelated variable")
	}
	data, err := os.ReadFile(filepath.Join(pluginDirectory, "mongodb.conf"))
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(homeDir, "zabbix-agent2-plugin", "mongodb.exe")
	if !strings.Contains(string(data), "Plugins.MongoDB.System.Path="+want) {
		t.Fatalf("MongoDB plugin path is missing from config: %s", data)
	}
}

func TestFeatureSwitchesRequireLowercaseTrue(t *testing.T) {
	env := bootstrap.Environment{
		"ZBX_ENABLEPERSISTENTBUFFER": "TRUE",
		"ZBX_PERSISTENTBUFFERFILE":   `C:\zabbix\buffer\agent2.db`,
		"ZBX_ENABLESTATUSPORT":       "false",
		"ZBX_STATUSPORT":             "31999",
	}
	configureFeatureSwitches(env)

	if _, found := env["ZBX_PERSISTENTBUFFERFILE"]; found {
		t.Fatal("persistent buffer file was retained")
	}
	if _, found := env["ZBX_ENABLEPERSISTENTBUFFER"]; found {
		t.Fatal("persistent buffer was enabled by a case-insensitive value")
	}
	if _, found := env["ZBX_STATUSPORT"]; found {
		t.Fatal("status port was retained")
	}
}
