//go:build windows

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

func TestPrepareService(t *testing.T) {
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
	configPath := filepath.Join(configDir, "zabbix_agent2.conf")
	config := "# Server=\n# ServerActive=\n# LogType=file\nLogFile=C:\\zabbix_agent2.log\nLogFileSize=1\n# EnablePersistentBuffer=0\n# PersistentBufferFile=\n# PersistentBufferPeriod=\nStatusPort=10000\n# Include=\n# Include=\n"
	if err := os.WriteFile(configPath, []byte(config), 0o600); err != nil {
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
		"ZBX_ENABLEPERSISTENTBUFFER": "true", "ZBX_PERSISTENTBUFFERPERIOD": "2h",
		"ZBX_SOURCEIP":       "192.0.2.1",
		"UNRELATED_VARIABLE": "value",
	}
	if err := prepareService(env); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	wants := []string{
		"Server=zabbix-server\n",
		"ServerActive=zabbix-server\n",
		"LogType=console\n",
		"SourceIP=192.0.2.1\n",
		"EnablePersistentBuffer=1\n",
		"PersistentBufferFile=" + filepath.Join(homeDir, "buffer", "agent2.db") + "\n",
		"PersistentBufferPeriod=2h\n",
		"StatusPort=12345\n",
		`Include=.\zabbix_agent2.d\plugins.d\*.conf` + "\n",
		`Include=.\zabbix_agentd.d\*.conf` + "\n",
	}
	for _, want := range wants {
		if !strings.Contains(content, want) {
			t.Fatalf("configuration does not contain %q:\n%s", want, data)
		}
	}
	if strings.Contains("\n"+content, "\nLogFile=") || strings.Contains("\n"+content, "\nLogFileSize=") {
		t.Fatalf("file logging parameters were not removed:\n%s", data)
	}
	if env["UNRELATED_VARIABLE"] != "value" {
		t.Fatal("entrypoint removed an unrelated variable")
	}
	if _, found := env["ZBX_STATUSPORT"]; found {
		t.Fatal("ZBX_STATUSPORT was not removed")
	}
	if env["ZABBIX_CONF_DIR"] != configDir {
		t.Fatal("ZABBIX_CONF_DIR was unexpectedly removed")
	}

	plugins := []struct {
		file, parameter, binary string
	}{
		{"mongodb.conf", "Plugins.MongoDB.System.Path", "mongodb.exe"},
		{"postgresql.conf", "Plugins.PostgreSQL.System.Path", "postgresql.exe"},
		{"mssql.conf", "Plugins.MSSQL.System.Path", "mssql.exe"},
		{"ember.conf", "Plugins.EmberPlus.System.Path", "ember-plus.exe"},
	}
	for _, plugin := range plugins {
		data, err = os.ReadFile(filepath.Join(pluginDirectory, plugin.file))
		if err != nil {
			t.Fatal(err)
		}
		want := plugin.parameter + "=" + filepath.Join(homeDir, "zabbix-agent2-plugin", plugin.binary)
		if !strings.Contains(string(data), want) {
			t.Fatalf("plugin path %q is missing from %s: %s", want, plugin.file, data)
		}
	}
}

func TestFeatureSwitchesAreCaseInsensitive(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "agent2.conf")
	if err := os.WriteFile(configPath, []byte("EnablePersistentBuffer=1\nPersistentBufferFile=old.db\nStatusPort=31999\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	env := bootstrap.Environment{
		"ZBX_ENABLEPERSISTENTBUFFER": "TRUE",
		"ZBX_PERSISTENTBUFFERFILE":   `C:\zabbix\buffer\agent2.db`,
		"ZBX_ENABLESTATUSPORT":       "false",
		"ZBX_STATUSPORT":             "31999",
	}
	if err := configureFeatureSwitches(env, t.TempDir(), configPath); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "EnablePersistentBuffer=1\n") {
		t.Fatalf("persistent buffer was not enabled by a case-insensitive value:\n%s", data)
	}
	if !strings.Contains(content, `PersistentBufferFile=C:\zabbix\buffer\agent2.db`+"\n") {
		t.Fatalf("configured persistent buffer path was not preserved:\n%s", data)
	}
	if strings.Contains(content, "StatusPort=") {
		t.Fatalf("status port was not removed:\n%s", data)
	}
}

func TestDisabledPersistentBufferClearsRelatedParameters(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "agent2.conf")
	config := "EnablePersistentBuffer=1\nPersistentBufferFile=old.db\nPersistentBufferPeriod=2h\n"
	if err := os.WriteFile(configPath, []byte(config), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := configureFeatureSwitches(bootstrap.Environment{}, t.TempDir(), configPath); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "EnablePersistentBuffer=0\n") {
		t.Fatalf("persistent buffer was not disabled:\n%s", data)
	}
	if strings.Contains(content, "PersistentBufferFile=") || strings.Contains(content, "PersistentBufferPeriod=") {
		t.Fatalf("disabled persistent buffer parameters were not removed:\n%s", data)
	}
}
