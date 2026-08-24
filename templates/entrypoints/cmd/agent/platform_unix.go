//go:build !windows

package main

import (
	"path/filepath"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/config"
)

const agentBinary = "/usr/sbin/zabbix_agentd"

func configurePerformanceCounters(bootstrap.Environment, string) error {
	return nil
}

func configureLoadModules(env bootstrap.Environment, configDir string) error {
	return config.MergeParameterValues(
		filepath.Join(configDir, "zabbix_agentd_modules.conf"),
		"LoadModule", env["ZBX_LOADMODULE"],
	)
}
