package main

import (
	"path/filepath"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

const agentBinary = `C:\zabbix\sbin\zabbix_agentd.exe`

func configurePerformanceCounters(env bootstrap.Environment, configDir string) error {
	configPath := filepath.Join(configDir, "zabbix_agentd_perf_counters.conf")

	if err := bootstrap.UpdateConfigIndexed(env, configPath, "PerfCounter", "ZBX_PERFCOUNTER"); err != nil {
		return err
	}

	return bootstrap.UpdateConfigIndexed(env, configPath, "PerfCounterEn", "ZBX_PERFCOUNTEREN")
}

// The Windows agent does not support loadable modules.
func configureLoadModules(bootstrap.Environment, string) error {
	return nil
}
