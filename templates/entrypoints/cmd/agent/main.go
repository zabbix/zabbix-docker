package main

import (
	config "github.com/zabbix/zabbix-docker/templates/entrypoints/internal/agent"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/hooks"
)

func prepareService(env bootstrap.Environment) error {
	bootstrap.LogInfo("** Preparing Zabbix agent")

	homeDir, configDir, err := bootstrap.RequiredDirectories(env)
	if err != nil {
		return err
	}

	config.ConfigureServers(env)

	if err := config.ConfigureAllowDenyKeys(env, configDir, "zabbix_agentd_item_keys.conf"); err != nil {
		return err
	}

	if err := configureLoadModules(env, configDir); err != nil {
		return err
	}

	if err := config.ProcessTLSFiles(env, homeDir); err != nil {
		return err
	}

	if err := bootstrap.ConfigureRunUser(env); err != nil {
		return err
	}

	if err := hooks.Run(env); err != nil {
		return err
	}

	config.ClearPrivateEnv(env)

	return nil
}

func main() {
	bootstrap.ExitOnError(bootstrap.RunService(agentBinary, prepareService))
}
