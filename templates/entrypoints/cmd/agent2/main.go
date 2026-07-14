package main

import (
	"os/exec"
	"path/filepath"

	agentconfig "github.com/zabbix/zabbix-docker/templates/entrypoints/internal/agent"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/hooks"
)

func prepareService(env bootstrap.Environment) error {
	bootstrap.LogInfo("** Preparing Zabbix agent 2")

	homeDirectory, configDirectory, err := bootstrap.RequiredDirectories(env)
	if err != nil {
		return err
	}

	agentconfig.ConfigureServers(env)
	configureFeatureSwitches(env)

	if err := agentconfig.ConfigureItemKeys(env, configDirectory, "zabbix_agent2_item_keys.conf"); err != nil {
		return err
	}

	if err := agentconfig.ProcessEncryptionFiles(env, homeDirectory); err != nil {
		return err
	}

	if err := updatePluginConfig(homeDirectory, configDirectory); err != nil {
		return err
	}

	if err := hooks.Run(env); err != nil {
		return err
	}

	agentconfig.ClearPrivateEnvironment(env)

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

func updatePluginConfig(homeDirectory, configDirectory string) error {
	bootstrap.LogInfo("** Preparing Zabbix agent 2 plugin configuration files")

	configDirectory = filepath.Join(configDirectory, "zabbix_agent2.d", "plugins.d")
	binDirectory := pluginBinaryDirectory(homeDirectory)

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
			filepath.Join(configDirectory, plugin.file),
			plugin.parameter,
			filepath.Join(binDirectory, plugin.binary+pluginExecutableSuffix),
		); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	bootstrap.ExitOnError(bootstrap.RunService(agent2Binary, prepareService))
}
