package main

import "github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"

const agentBinary = `C:\zabbix\sbin\zabbix_agentd.exe`

// The Windows agent does not support loadable modules.
func configureLoadModules(bootstrap.Environment, string) error {
	return nil
}
