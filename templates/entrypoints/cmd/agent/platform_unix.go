//go:build !windows

package main

import (
	"path/filepath"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

const agentBinary = "/usr/sbin/zabbix_agentd"

func configureLoadModules(env bootstrap.Environment, configDirectory string) error {
	return bootstrap.UpdateConfigMultiple(
		filepath.Join(configDirectory, "zabbix_agentd_modules.conf"),
		"LoadModule",
		env["ZBX_LOADMODULE"],
	)
}
