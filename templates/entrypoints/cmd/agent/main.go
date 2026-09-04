//go:build windows

package main

import (
	"path/filepath"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/config"
)

const agentBinary = `C:\zabbix\sbin\zabbix_agentd.exe`

func prepareService(env bootstrap.Environment) error {
	bootstrap.LogInfo("** Preparing Zabbix agent")

	homeDir, configDir, err := bootstrap.CommonDirs(env)
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

	if err := config.SetParameters(configPath,
		config.Parameter{Name: "LogType", Value: "console"},
		config.Parameter{Name: "LogFile"},
		config.Parameter{Name: "LogFileSize"},

		config.Parameter{Name: "DebugLevel", Value: env["ZBX_DEBUGLEVEL"]},

		config.Parameter{Name: "SourceIP", Value: env["ZBX_SOURCEIP"]},

		config.Parameter{Name: "EnableRemoteCommands", Value: env["ZBX_ENABLEREMOTECOMMANDS"]},
		config.Parameter{Name: "LogRemoteCommands", Value: env["ZBX_LOGREMOTECOMMANDS"]},

		config.Parameter{Name: "ListenPort", Value: env["ZBX_LISTENPORT"]},
		config.Parameter{Name: "ListenIP", Value: env["ZBX_LISTENIP"]},

		config.Parameter{Name: "StartAgents", Value: env["ZBX_STARTAGENTS"]},
		config.Parameter{Name: "HeartbeatFrequency", Value: env["ZBX_HEARTBEAT_FREQUENCY"]},

		config.Parameter{Name: "HostInterface", Value: env["ZBX_HOSTINTERFACE"]},
		config.Parameter{Name: "HostInterfaceItem", Value: env["ZBX_HOSTINTERFACEITEM"]},
		config.Parameter{Name: "Hostname", Value: env["ZBX_HOSTNAME"]},
		config.Parameter{Name: "HostnameItem", Value: env["ZBX_HOSTNAMEITEM"]},
		config.Parameter{Name: "HostMetadata", Value: env["ZBX_METADATA"]},
		config.Parameter{Name: "HostMetadataItem", Value: env["ZBX_METADATAITEM"]},

		config.Parameter{Name: "RefreshActiveChecks", Value: env["ZBX_REFRESHACTIVECHECKS"]},

		config.Parameter{Name: "BufferSend", Value: env["ZBX_BUFFERSEND"]},
		config.Parameter{Name: "BufferSize", Value: env["ZBX_BUFFERSIZE"]},

		config.Parameter{Name: "MaxLinesPerSecond", Value: env["ZBX_MAXLINESPERSECOND"]},
		config.Parameter{Name: "Timeout", Value: env["ZBX_TIMEOUT"]},
		config.Parameter{Name: "Include", Value: filepath.Join(configDir, "zabbix_agentd.d", "*.conf")},

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

	if err := config.ConfigureAllowDenyKeys(env, configPath); err != nil {
		return err
	}

	config.ClearPrivateEnv(env)

	return nil
}

func main() {
	bootstrap.ExitOnError(bootstrap.RunService(agentBinary, prepareService))
}
