package main

import (
	"path/filepath"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/config"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/hooks"
)

func prepareService(env bootstrap.Environment) error {
	bootstrap.LogInfo("** Preparing Zabbix agent")

	homeDir, configDir, err := bootstrap.CommonDirs(env)
	if err != nil {
		return err
	}

	config.ConfigureServers(env)

	if err := config.ConfigureItemKeyRules(env, configDir, "zabbix_agentd_item_keys.conf"); err != nil {
		return err
	}

	if err := config.UpdateIndexedParameter(env, filepath.Join(configDir, "zabbix_agentd_aliases.conf"), "Alias", "ZBX_ALIAS"); err != nil {
		return err
	}

	if err := config.UpdateIndexedParameter(env, filepath.Join(configDir, "zabbix_agentd_user_parameters.conf"), "UserParameter", "ZBX_USERPARAMETER"); err != nil {
		return err
	}

	if err := configurePerformanceCounters(env, configDir); err != nil {
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
	bootstrap.Main(bootstrap.Service(agentBinary, prepareService))
}
