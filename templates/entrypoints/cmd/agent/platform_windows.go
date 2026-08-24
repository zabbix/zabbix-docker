package main

import (
	"path/filepath"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/config"
)

const agentBinary = `C:\zabbix\sbin\zabbix_agentd.exe`

func configurePerformanceCounters(env bootstrap.Environment, configDir string) error {
	configPath := filepath.Join(configDir, "zabbix_agentd_perf_counters.conf")

	if err := config.UpdateIndexedParameter(env, configPath, "PerfCounter", "ZBX_PERFCOUNTER"); err != nil {
		return err
	}

	return config.UpdateIndexedParameter(env, configPath, "PerfCounterEn", "ZBX_PERFCOUNTEREN")
}

// Windows agent does not support loadable modules.
func configureLoadModules(bootstrap.Environment, string) error {
	return nil
}
