package main

import (
	agentconfig "github.com/zabbix/zabbix-docker/templates/entrypoints/internal/agent"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/hooks"
)

func prepareService(env bootstrap.Environment) error {
	bootstrap.LogInfo("** Preparing Zabbix agent")

	homeDirectory, configDirectory, err := bootstrap.RequiredDirectories(env)
	if err != nil {
		return err
	}

	agentconfig.ConfigureServers(env)

	if err := agentconfig.ConfigureItemKeys(env, configDirectory, "zabbix_agentd_item_keys.conf"); err != nil {
		return err
	}

	if err := configureLoadModules(env, configDirectory); err != nil {
		return err
	}

	if err := agentconfig.ProcessEncryptionFiles(env, homeDirectory); err != nil {
		return err
	}

	if err := bootstrap.ConfigureIdentity(env); err != nil {
		return err
	}

	if err := hooks.Run(env); err != nil {
		return err
	}

	agentconfig.ClearPrivateEnvironment(env)

	return nil
}

func main() {
	bootstrap.ExitOnError(bootstrap.RunService(agentBinary, prepareService))
}
