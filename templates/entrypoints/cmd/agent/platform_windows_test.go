package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

func TestPrepareServiceWindows(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "conf")
	homeDir := filepath.Join(root, "home")
	if err := os.MkdirAll(filepath.Join(homeDir, "enc_internal"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string]string{
		"zabbix_agentd_aliases.conf":       "# Alias=\n",
		"zabbix_agentd_item_keys.conf":     "# DenyKey=system.run[*]\n",
		"zabbix_agentd_perf_counters.conf": "# PerfCounter=\n# PerfCounterEn=\n",
	} {
		if err := os.WriteFile(filepath.Join(configDir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	env := bootstrap.Environment{
		"ZABBIX_CONF_DIR": configDir, "ZABBIX_USER_HOME_DIR": homeDir,
		"ZBX_ALIAS_0":         "custom.echo:system.localtime",
		"ZBX_PERFCOUNTER_0":   `interrupts,"\Processor(0)\Interrupts/sec",60`,
		"ZBX_PERFCOUNTEREN_0": `interrupts.en,"\Processor(0)\Interrupts/sec",60`,
		"MYSQL_PASSWORD":      "password",
	}
	if err := prepareService(env); err != nil {
		t.Fatal(err)
	}
	if env["ZBX_PASSIVESERVERS"] != "zabbix-server" || env["ZBX_ACTIVESERVERS"] != "zabbix-server" {
		t.Fatalf("unexpected server configuration: %#v", env)
	}
	if env["MYSQL_PASSWORD"] != "password" {
		t.Fatal("Windows entrypoint removed a non-ZABBIX variable")
	}
	if _, found := env["ZABBIX_CONF_DIR"]; found {
		t.Fatal("ZABBIX_CONF_DIR was not removed")
	}
	aliases, err := os.ReadFile(filepath.Join(configDir, "zabbix_agentd_aliases.conf"))
	if err != nil || !strings.Contains(string(aliases), "Alias=${ZBX_ALIAS_0}") {
		t.Fatalf("alias config: %s, %v", aliases, err)
	}
	performanceCounters, err := os.ReadFile(filepath.Join(configDir, "zabbix_agentd_perf_counters.conf"))
	if err != nil || !strings.Contains(string(performanceCounters), "PerfCounter=${ZBX_PERFCOUNTER_0}") ||
		!strings.Contains(string(performanceCounters), "PerfCounterEn=${ZBX_PERFCOUNTEREN_0}") {
		t.Fatalf("performance counter config: %s, %v", performanceCounters, err)
	}
}
