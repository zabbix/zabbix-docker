//go:build windows

package main

import (
	"path/filepath"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/config"
)

const agent2Binary = `C:\zabbix\sbin\zabbix_agent2.exe`

func prepareService(env bootstrap.Environment) error {
	bootstrap.LogInfo("** Preparing Zabbix agent 2")

	homeDir, configDir, err := bootstrap.CommonDirs(env)
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

	if err := config.SetParameters(configPath,
		config.Parameter{Name: "LogType", Value: "console"},
		config.Parameter{Name: "LogFile"},
		config.Parameter{Name: "LogFileSize"},

		config.Parameter{Name: "DebugLevel", Value: env["ZBX_DEBUGLEVEL"]},

		config.Parameter{Name: "SourceIP", Value: env["ZBX_SOURCEIP"]},

		config.Parameter{Name: "Plugins.SystemRun.LogRemoteCommands", Value: env["ZBX_LOGREMOTECOMMANDS"]},
		config.Parameter{Name: "ListenPort", Value: env["ZBX_LISTENPORT"]},
		config.Parameter{Name: "ListenIP", Value: env["ZBX_LISTENIP"]},
		config.Parameter{Name: "HeartbeatFrequency", Value: env["ZBX_HEARTBEAT_FREQUENCY"]},
		config.Parameter{Name: "ForceActiveChecksOnStart", Value: env["ZBX_FORCEACTIVECHECKSONSTART"]},

		config.Parameter{Name: "Hostname", Value: env["ZBX_HOSTNAME"]},
		config.Parameter{Name: "HostnameItem", Value: env["ZBX_HOSTNAMEITEM"]},
		config.Parameter{Name: "HostMetadata", Value: env["ZBX_METADATA"]},
		config.Parameter{Name: "HostMetadataItem", Value: env["ZBX_METADATAITEM"]},
		config.Parameter{Name: "HostInterface", Value: env["ZBX_HOSTINTERFACE"]},
		config.Parameter{Name: "HostInterfaceItem", Value: env["ZBX_HOSTINTERFACEITEM"]},

		config.Parameter{Name: "RefreshActiveChecks", Value: env["ZBX_REFRESHACTIVECHECKS"]},

		config.Parameter{Name: "BufferSend", Value: env["ZBX_BUFFERSEND"]},
		config.Parameter{Name: "BufferSize", Value: env["ZBX_BUFFERSIZE"]},

		config.Parameter{Name: "Plugins.Log.MaxLinesPerSecond", Value: env["ZBX_MAXLINESPERSECOND"]},
		config.Parameter{Name: "Plugins.EventLog.MaxLinesPerSecond", Value: env["ZBX_EVENTLOGMAXLINESPERSECOND"]},
		config.Parameter{Name: "PluginTimeout", Value: env["ZBX_PLUGINTIMEOUT"]},
		config.Parameter{Name: "Timeout", Value: env["ZBX_TIMEOUT"]},

		config.Parameter{Name: "UnsafeUserParameters", Value: env["ZBX_UNSAFEUSERPARAMETERS"]},
		config.Parameter{Name: "UserParameterDir", Value: env["ZBX_USERPARAMETERDIR"]},

		config.Parameter{Name: "TLSConnect", Value: env["ZBX_TLSCONNECT"]},
		config.Parameter{Name: "TLSAccept", Value: env["ZBX_TLSACCEPT"]},
		config.Parameter{Name: "TLSServerCertIssuer", Value: env["ZBX_TLSSERVERCERTISSUER"]},
		config.Parameter{Name: "TLSServerCertSubject", Value: env["ZBX_TLSSERVERCERTSUBJECT"]},

		config.Parameter{Name: "TLSCipherAll", Value: env["ZBX_TLSCIPHERALL"]},
		config.Parameter{Name: "TLSCipherAll13", Value: env["ZBX_TLSCIPHERALL13"]},
		config.Parameter{Name: "TLSCipherCert", Value: env["ZBX_TLSCIPHERCERT"]},
		config.Parameter{Name: "TLSCipherCert13", Value: env["ZBX_TLSCIPHERCERT13"]},
		config.Parameter{Name: "TLSCipherPSK", Value: env["ZBX_TLSCIPHERPSK"]},
		config.Parameter{Name: "TLSCipherPSK13", Value: env["ZBX_TLSCIPHERPSK13"]},
		config.Parameter{Name: "TLSPSKIdentity", Value: env["ZBX_TLSPSKIDENTITY"]},
	); err != nil {
		return err
	}

	if err := configureFeatureSwitches(env, homeDir, configPath); err != nil {
		return err
	}

	if err := config.SetParameter(configPath, "Include", `.\zabbix_agent2.d\plugins.d\*.conf`); err != nil {
		return err
	}
	if err := config.MergeParameterValues(configPath, "Include", `.\zabbix_agentd.d\*.conf`); err != nil {
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

// configureFeatureSwitches updates Agent 2-specific persistent buffer and
// status listener settings.
func configureFeatureSwitches(env bootstrap.Environment, homeDir, configPath string) error {
	values := []config.Parameter{}

	if env["ZBX_ENABLEPERSISTENTBUFFER"] == "true" {
		values = append(values,
			config.Parameter{Name: "EnablePersistentBuffer", Value: "1"},
			config.Parameter{
				Name:  "PersistentBufferFile",
				Value: env.ValueOrDefaultNonEmpty("ZBX_PERSISTENTBUFFERFILE", filepath.Join(homeDir, "buffer", "agent2.db")),
			},
			config.Parameter{Name: "PersistentBufferPeriod", Value: env["ZBX_PERSISTENTBUFFERPERIOD"]},
		)
	} else {
		values = append(values, config.Parameter{Name: "EnablePersistentBuffer", Value: "0"})
	}

	if env["ZBX_ENABLESTATUSPORT"] == "true" {
		values = append(values, config.Parameter{
			Name:  "StatusPort",
			Value: env.ValueOrDefaultNonEmpty("ZBX_STATUSPORT", "31999"),
		})
	} else {
		values = append(values, config.Parameter{Name: "StatusPort"})
	}

	return config.SetParameters(configPath, values...)
}

// updatePluginConfig writes platform-specific executable paths for Agent 2
// external plugins.
func updatePluginConfig(homeDir, configDir string) error {
	bootstrap.LogInfo("** Preparing Zabbix agent 2 plugin configuration files")

	configDir = filepath.Join(configDir, "zabbix_agent2.d", "plugins.d")
	binDir := filepath.Join(homeDir, "zabbix-agent2-plugin")

	plugins := []struct {
		file, param, binary string
	}{
		{"mongodb.conf", "Plugins.MongoDB.System.Path", "mongodb"},
		{"postgresql.conf", "Plugins.PostgreSQL.System.Path", "postgresql"},
		{"mssql.conf", "Plugins.MSSQL.System.Path", "mssql"},
		{"ember.conf", "Plugins.EmberPlus.System.Path", "ember-plus"},
	}

	for _, plugin := range plugins {
		if err := config.SetParameter(
			filepath.Join(configDir, plugin.file),
			plugin.param,
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
