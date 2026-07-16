package main

import (
	"path/filepath"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

const (
	agent2Binary     = `C:\zabbix\sbin\zabbix_agent2.exe`
	nvidiaCommand    = "nvidia-smi.exe"
	pluginExecSuffix = ".exe"
)

func configurePerformanceCounters(env bootstrap.Environment, configDir string) error {
	configPath := filepath.Join(configDir, "zabbix_agent2_perf_counters.conf")

	if err := bootstrap.UpdateConfigIndexed(env, configPath, "PerfCounter", "ZBX_PERFCOUNTER"); err != nil {
		return err
	}

	return bootstrap.UpdateConfigIndexed(env, configPath, "PerfCounterEn", "ZBX_PERFCOUNTEREN")
}

func pluginBinDir(homeDir string) string {
	return filepath.Join(homeDir, "zabbix-agent2-plugin")
}
