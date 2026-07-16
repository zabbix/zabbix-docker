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
	for name, content := range map[string]string{
		"zabbix_agent2_aliases.conf":       "# Alias=\n",
		"zabbix_agent2_item_keys.conf":     "# DenyKey=system.run[*]\n",
		"zabbix_agent2_perf_counters.conf": "# PerfCounter=\n# PerfCounterEn=\n",
	} {
		if err := os.WriteFile(filepath.Join(configDir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"mongodb.conf", "postgresql.conf", "mssql.conf", "ember.conf"} {
		if err := os.WriteFile(filepath.Join(pluginDirectory, name), []byte("# plugin config\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	env := bootstrap.Environment{
		"ZABBIX_CONF_DIR": configDir, "ZABBIX_USER_HOME_DIR": homeDir,
		"ZBX_ENABLESTATUSPORT": "true", "ZBX_STATUSPORT": "12345",
		"ZBX_ALIAS_0":         "custom.echo:system.localtime",
		"ZBX_PERFCOUNTER_0":   `interrupts,"\Processor(0)\Interrupts/sec",60`,
		"ZBX_PERFCOUNTEREN_0": `interrupts.en,"\Processor(0)\Interrupts/sec",60`,
		"POSTGRES_PASSWORD":   "password",
	}
	if err := prepareService(env); err != nil {
		t.Fatal(err)
	}
	if env["ZBX_STATUSPORT"] != "12345" {
		t.Fatalf("unexpected Windows status port: %q", env["ZBX_STATUSPORT"])
	}
	if env["POSTGRES_PASSWORD"] != "password" {
		t.Fatal("Windows entrypoint removed a non-ZABBIX variable")
	}
	data, err := os.ReadFile(filepath.Join(pluginDirectory, "mongodb.conf"))
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(homeDir, "zabbix-agent2-plugin", "mongodb.exe")
	if !strings.Contains(string(data), "Plugins.MongoDB.System.Path="+want) {
		t.Fatalf("MongoDB plugin path is missing from config: %s", data)
	}
	aliases, err := os.ReadFile(filepath.Join(configDir, "zabbix_agent2_aliases.conf"))
	if err != nil || !strings.Contains(string(aliases), "Alias=${ZBX_ALIAS_0}") {
		t.Fatalf("alias config: %s, %v", aliases, err)
	}
	performanceCounters, err := os.ReadFile(filepath.Join(configDir, "zabbix_agent2_perf_counters.conf"))
	if err != nil || !strings.Contains(string(performanceCounters), "PerfCounter=${ZBX_PERFCOUNTER_0}") ||
		!strings.Contains(string(performanceCounters), "PerfCounterEn=${ZBX_PERFCOUNTEREN_0}") {
		t.Fatalf("performance counter config: %s, %v", performanceCounters, err)
	}
}
