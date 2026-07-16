//go:build windows

package main

import (
	"path/filepath"

	config "github.com/zabbix/zabbix-docker/templates/entrypoints/internal/agent"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

const agentBinary = `C:\zabbix\sbin\zabbix_agentd.exe`

func prepareService(env bootstrap.Environment) error {
	bootstrap.LogInfo("** Preparing Zabbix agent")

	homeDir, configDir, err := bootstrap.RequiredDirectories(env)
	if err != nil {
		return err
	}
	configPath := filepath.Join(configDir, "zabbix_agentd.conf")

	bootstrap.LogInfo("** Preparing Zabbix agent configuration file")

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

		bootstrap.ConfigValue{Name: "EnableRemoteCommands", Value: env["ZBX_ENABLEREMOTECOMMANDS"]},
		bootstrap.ConfigValue{Name: "LogRemoteCommands", Value: env["ZBX_LOGREMOTECOMMANDS"]},

		bootstrap.ConfigValue{Name: "ListenPort", Value: env["ZBX_LISTENPORT"]},
		bootstrap.ConfigValue{Name: "ListenIP", Value: env["ZBX_LISTENIP"]},

		bootstrap.ConfigValue{Name: "StartAgents", Value: env["ZBX_STARTAGENTS"]},
		bootstrap.ConfigValue{Name: "HeartbeatFrequency", Value: env["ZBX_HEARTBEAT_FREQUENCY"]},

		bootstrap.ConfigValue{Name: "HostInterface", Value: env["ZBX_HOSTINTERFACE"]},
		bootstrap.ConfigValue{Name: "HostInterfaceItem", Value: env["ZBX_HOSTINTERFACEITEM"]},
		bootstrap.ConfigValue{Name: "Hostname", Value: env["ZBX_HOSTNAME"]},
		bootstrap.ConfigValue{Name: "HostnameItem", Value: env["ZBX_HOSTNAMEITEM"]},
		bootstrap.ConfigValue{Name: "HostMetadata", Value: env["ZBX_METADATA"]},
		bootstrap.ConfigValue{Name: "HostMetadataItem", Value: env["ZBX_METADATAITEM"]},

		bootstrap.ConfigValue{Name: "RefreshActiveChecks", Value: env["ZBX_REFRESHACTIVECHECKS"]},

		bootstrap.ConfigValue{Name: "BufferSend", Value: env["ZBX_BUFFERSEND"]},
		bootstrap.ConfigValue{Name: "BufferSize", Value: env["ZBX_BUFFERSIZE"]},

		bootstrap.ConfigValue{Name: "MaxLinesPerSecond", Value: env["ZBX_MAXLINESPERSECOND"]},
		bootstrap.ConfigValue{Name: "Timeout", Value: env["ZBX_TIMEOUT"]},
		bootstrap.ConfigValue{Name: "Include", Value: filepath.Join(configDir, "zabbix_agentd.d", "*.conf")},

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

	if err := config.ConfigureAllowDenyKeys(env, configPath); err != nil {
		return err
	}

	config.ClearPrivateEnv(env)

	return nil
}

func main() {
	bootstrap.ExitOnError(bootstrap.RunService(agentBinary, prepareService))
}
