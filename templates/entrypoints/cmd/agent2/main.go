//go:build windows

package main

import (
	"path/filepath"

	config "github.com/zabbix/zabbix-docker/templates/entrypoints/internal/agent"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

const agent2Binary = `C:\zabbix\sbin\zabbix_agent2.exe`

func prepareService(env bootstrap.Environment) error {
	bootstrap.LogInfo("** Preparing Zabbix agent 2")

	homeDir, configDir, err := bootstrap.RequiredDirectories(env)
	if err != nil {
		return err
	}
	configPath := filepath.Join(configDir, "zabbix_agent2.conf")

	bootstrap.LogInfo("** Preparing Zabbix agent 2 configuration file")

	if err := config.ConfigureServers(env, configPath); err != nil {
		return err
	}
	if err := config.ProcessTLSFiles(env, homeDir, configPath); err != nil {
		return err
	}

	if err := bootstrap.UpdateConfigValues(configPath,
		bootstrap.ConfigValue{Name: "LogType", Value: "console"},
		bootstrap.ConfigValue{Name: "LogFile"},
		bootstrap.ConfigValue{Name: "LogFileSize"},

		bootstrap.ConfigValue{Name: "DebugLevel", Value: env["ZBX_DEBUGLEVEL"]},

		bootstrap.ConfigValue{Name: "SourceIP", Value: env["ZBX_SOURCEIP"]},

		bootstrap.ConfigValue{Name: "Plugins.SystemRun.LogRemoteCommands", Value: env["ZBX_LOGREMOTECOMMANDS"]},
		bootstrap.ConfigValue{Name: "ListenPort", Value: env["ZBX_LISTENPORT"]},
		bootstrap.ConfigValue{Name: "ListenIP", Value: env["ZBX_LISTENIP"]},
		bootstrap.ConfigValue{Name: "HeartbeatFrequency", Value: env["ZBX_HEARTBEAT_FREQUENCY"]},
		bootstrap.ConfigValue{Name: "ForceActiveChecksOnStart", Value: env["ZBX_FORCEACTIVECHECKSONSTART"]},

		bootstrap.ConfigValue{Name: "Hostname", Value: env["ZBX_HOSTNAME"]},
		bootstrap.ConfigValue{Name: "HostnameItem", Value: env["ZBX_HOSTNAMEITEM"]},
		bootstrap.ConfigValue{Name: "HostMetadata", Value: env["ZBX_METADATA"]},
		bootstrap.ConfigValue{Name: "HostMetadataItem", Value: env["ZBX_METADATAITEM"]},
		bootstrap.ConfigValue{Name: "HostInterface", Value: env["ZBX_HOSTINTERFACE"]},
		bootstrap.ConfigValue{Name: "HostInterfaceItem", Value: env["ZBX_HOSTINTERFACEITEM"]},

		bootstrap.ConfigValue{Name: "RefreshActiveChecks", Value: env["ZBX_REFRESHACTIVECHECKS"]},

		bootstrap.ConfigValue{Name: "BufferSend", Value: env["ZBX_BUFFERSEND"]},
		bootstrap.ConfigValue{Name: "BufferSize", Value: env["ZBX_BUFFERSIZE"]},

		bootstrap.ConfigValue{Name: "Plugins.Log.MaxLinesPerSecond", Value: env["ZBX_MAXLINESPERSECOND"]},
		bootstrap.ConfigValue{Name: "Plugins.EventLog.MaxLinesPerSecond", Value: env["ZBX_EVENTLOGMAXLINESPERSECOND"]},
		bootstrap.ConfigValue{Name: "PluginTimeout", Value: env["ZBX_PLUGINTIMEOUT"]},
		bootstrap.ConfigValue{Name: "Timeout", Value: env["ZBX_TIMEOUT"]},

		bootstrap.ConfigValue{Name: "UnsafeUserParameters", Value: env["ZBX_UNSAFEUSERPARAMETERS"]},
		bootstrap.ConfigValue{Name: "UserParameterDir", Value: env["ZBX_USERPARAMETERDIR"]},

		bootstrap.ConfigValue{Name: "TLSConnect", Value: env["ZBX_TLSCONNECT"]},
		bootstrap.ConfigValue{Name: "TLSAccept", Value: env["ZBX_TLSACCEPT"]},
		bootstrap.ConfigValue{Name: "TLSServerCertIssuer", Value: env["ZBX_TLSSERVERCERTISSUER"]},
		bootstrap.ConfigValue{Name: "TLSServerCertSubject", Value: env["ZBX_TLSSERVERCERTSUBJECT"]},

		bootstrap.ConfigValue{Name: "TLSCipherAll", Value: env["ZBX_TLSCIPHERALL"]},
		bootstrap.ConfigValue{Name: "TLSCipherAll13", Value: env["ZBX_TLSCIPHERALL13"]},
		bootstrap.ConfigValue{Name: "TLSCipherCert", Value: env["ZBX_TLSCIPHERCERT"]},
		bootstrap.ConfigValue{Name: "TLSCipherCert13", Value: env["ZBX_TLSCIPHERCERT13"]},
		bootstrap.ConfigValue{Name: "TLSCipherPSK", Value: env["ZBX_TLSCIPHERPSK"]},
		bootstrap.ConfigValue{Name: "TLSCipherPSK13", Value: env["ZBX_TLSCIPHERPSK13"]},
		bootstrap.ConfigValue{Name: "TLSPSKIdentity", Value: env["ZBX_TLSPSKIDENTITY"]},
	); err != nil {
		return err
	}

	if err := configureFeatureSwitches(env, homeDir, configPath); err != nil {
		return err
	}

	if err := bootstrap.UpdateConfigValue(configPath, "Include", `.\zabbix_agent2.d\plugins.d\*.conf`); err != nil {
		return err
	}
	if err := bootstrap.UpdateConfigMultiple(configPath, "Include", `.\zabbix_agentd.d\*.conf`); err != nil {
		return err
	}

	if err := config.ConfigureAllowDenyKeys(env, configPath); err != nil {
		return err
	}

	if err := updatePluginConfig(homeDir, configDir); err != nil {
		return err
	}

	config.ClearPrivateEnv(env)

	return nil
}

func configureFeatureSwitches(env bootstrap.Environment, homeDir, configPath string) error {
	values := []bootstrap.ConfigValue{}

	if env["ZBX_ENABLEPERSISTENTBUFFER"] == "true" {
		values = append(values,
			bootstrap.ConfigValue{Name: "EnablePersistentBuffer", Value: "1"},
			bootstrap.ConfigValue{
				Name:  "PersistentBufferFile",
				Value: env.ValueOrDefaultNonEmpty("ZBX_PERSISTENTBUFFERFILE", filepath.Join(homeDir, "buffer", "agent2.db")),
			},
			bootstrap.ConfigValue{Name: "PersistentBufferPeriod", Value: env["ZBX_PERSISTENTBUFFERPERIOD"]},
		)
	} else {
		values = append(values, bootstrap.ConfigValue{Name: "EnablePersistentBuffer", Value: "0"})
	}

	if env["ZBX_ENABLESTATUSPORT"] == "true" {
		values = append(values, bootstrap.ConfigValue{
			Name:  "StatusPort",
			Value: env.ValueOrDefaultNonEmpty("ZBX_STATUSPORT", "31999"),
		})
	} else {
		values = append(values, bootstrap.ConfigValue{Name: "StatusPort"})
	}

	return bootstrap.UpdateConfigValues(configPath, values...)
}

func updatePluginConfig(homeDir, configDir string) error {
	bootstrap.LogInfo("** Preparing Zabbix agent 2 plugin configuration files")

	configDir = filepath.Join(configDir, "zabbix_agent2.d", "plugins.d")
	binDir := filepath.Join(homeDir, "zabbix-agent2-plugin")

	plugins := []struct {
		file, parameter, binary string
	}{
		{"mongodb.conf", "Plugins.MongoDB.System.Path", "mongodb"},
		{"postgresql.conf", "Plugins.PostgreSQL.System.Path", "postgresql"},
		{"mssql.conf", "Plugins.MSSQL.System.Path", "mssql"},
		{"ember.conf", "Plugins.EmberPlus.System.Path", "ember-plus"},
	}

	for _, plugin := range plugins {
		if err := bootstrap.UpdateConfigValue(
			filepath.Join(configDir, plugin.file),
			plugin.parameter,
			filepath.Join(binDir, plugin.binary+".exe"),
		); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	bootstrap.ExitOnError(bootstrap.RunService(agent2Binary, prepareService))
}
