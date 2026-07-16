package main

import (
	"os/exec"
	"path/filepath"

	config "github.com/zabbix/zabbix-docker/templates/entrypoints/internal/agent"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/hooks"
)

func prepareService(env bootstrap.Environment) error {
	bootstrap.LogInfo("** Preparing Zabbix agent 2")

	homeDir, configDir, err := bootstrap.RequiredDirectories(env)
	if err != nil {
		return err
	}

	config.ConfigureServers(env)
	configureFeatureSwitches(env)

	if err := config.ConfigureItemKeyRules(env, configDir, "zabbix_agent2_item_keys.conf"); err != nil {
		return err
	}

	if err := bootstrap.UpdateConfigIndexed(env, filepath.Join(configDir, "zabbix_agent2_aliases.conf"), "Alias", "ZBX_ALIAS"); err != nil {
		return err
	}

	if err := bootstrap.UpdateConfigIndexed(env, filepath.Join(configDir, "zabbix_agent2_user_parameters.conf"), "UserParameter", "ZBX_USERPARAMETER"); err != nil {
		return err
	}

	if err := configurePerformanceCounters(env, configDir); err != nil {
		return err
	}

	if err := config.ProcessTLSFiles(env, homeDir); err != nil {
		return err
	}

	if err := updatePluginConfig(homeDir, configDir); err != nil {
		return err
	}

	if err := hooks.Run(env); err != nil {
		return err
	}

	config.ClearPrivateEnv(env)

	return nil
}

func configureFeatureSwitches(env bootstrap.Environment) {
	if env["ZBX_ENABLEPERSISTENTBUFFER"] == "true" {
		env["ZBX_ENABLEPERSISTENTBUFFER"] = "1"
	} else {
		delete(env, "ZBX_ENABLEPERSISTENTBUFFER")
		delete(env, "ZBX_PERSISTENTBUFFERFILE")
	}

	if env["ZBX_ENABLESTATUSPORT"] == "true" {
		env["ZBX_STATUSPORT"] = env.ValueOrDefaultNonEmpty("ZBX_STATUSPORT", "31999")
	} else {
		delete(env, "ZBX_STATUSPORT")
	}
}

func updatePluginConfig(homeDir, configDir string) error {
	bootstrap.LogInfo("** Preparing Zabbix agent 2 plugin configuration files")

	configDir = filepath.Join(configDir, "zabbix_agent2.d", "plugins.d")
	binDir := pluginBinDir(homeDir)

	plugins := []struct {
		file, parameter, binary string
	}{
		{"mongodb.conf", "Plugins.MongoDB.System.Path", "mongodb"},
		{"postgresql.conf", "Plugins.PostgreSQL.System.Path", "postgresql"},
		{"mssql.conf", "Plugins.MSSQL.System.Path", "mssql"},
		{"ember.conf", "Plugins.EmberPlus.System.Path", "ember-plus"},
	}

	if _, err := exec.LookPath(nvidiaCommand); err == nil {
		plugins = append(plugins, struct {
			file, parameter, binary string
		}{"nvidia.conf", "Plugins.NVIDIA.System.Path", "nvidia-gpu"})
	}

	for _, plugin := range plugins {
		if err := bootstrap.UpdateConfigValue(
			filepath.Join(configDir, plugin.file),
			plugin.parameter,
			filepath.Join(binDir, plugin.binary+pluginExecSuffix),
		); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	bootstrap.ExitOnError(bootstrap.RunService(agent2Binary, prepareService))
}
