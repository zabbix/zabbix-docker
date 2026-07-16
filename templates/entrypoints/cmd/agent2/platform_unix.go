//go:build !windows

package main

import "github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"

const (
	agent2Binary     = "/usr/sbin/zabbix_agent2"
	nvidiaCommand    = "nvidia-smi"
	pluginExecSuffix = ""
)

func configurePerformanceCounters(bootstrap.Environment, string) error {
	return nil
}

func pluginBinDir(homeDir string) string {
	return "/usr/sbin/zabbix-agent2-plugin"
}
