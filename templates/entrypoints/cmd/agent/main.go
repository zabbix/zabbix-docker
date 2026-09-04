//go:build windows

package main

import (
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/config"
)

func prepareService(env bootstrap.Environment) error {
	bootstrap.LogInfo("** Preparing Zabbix agent")

	homeDir, configDir, err := bootstrap.CommonDirs(env)
	if err != nil {
		return err
	}

	config.ConfigureServers(env)

	if err := config.ConfigureAllowDenyKeys(env, configDir, "zabbix_agentd_item_keys.conf"); err != nil {
		return err
	}

	if err := config.ProcessTLSFiles(env, homeDir); err != nil {
		return err
	}

	config.ClearPrivateEnv(env)

	return nil
}

func main() {
	bootstrap.Main(bootstrap.Service(agentBinary, prepareService))
}
