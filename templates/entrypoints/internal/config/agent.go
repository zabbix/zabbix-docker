//go:build windows

// Package config prepares Zabbix configuration from the container environment.
package config

import (
	"strings"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

// ConfigureServers merges ZBX_SERVER_HOST and ZBX_SERVER_PORT into the
// passive and active server lists and writes them to configPath, honouring
// the ZBX_PASSIVE_ALLOW and ZBX_ACTIVE_ALLOW switches.
func ConfigureServers(env bootstrap.Environment, configPath string) error {
	serverHost := env["ZBX_SERVER_HOST"]
	if strings.TrimSpace(serverHost) == "" {
		serverHost = "zabbix-server"
	}
	serverPort := env["ZBX_SERVER_PORT"]
	if strings.TrimSpace(serverPort) == "" {
		serverPort = "10051"
	}
	passiveServers := env["ZBX_PASSIVESERVERS"]
	activeServers := env["ZBX_ACTIVESERVERS"]

	activeServer := serverHost
	if serverPort != "" && serverPort != "10051" {
		activeServer += ":" + serverPort
	}
	if serverHost != "" {
		passiveServers = prependServer(serverHost, passiveServers)
		activeServers = prependServer(activeServer, activeServers)
	}

	if value := env["ZBX_PASSIVE_ALLOW"]; (strings.TrimSpace(value) == "" || strings.EqualFold(value, "true")) && passiveServers != "" {
		bootstrap.LogInfo("** Using '%s' servers for passive checks", passiveServers)
		env["ZBX_PASSIVESERVERS"] = passiveServers
	} else {
		delete(env, "ZBX_PASSIVESERVERS")
	}

	if value := env["ZBX_ACTIVE_ALLOW"]; (strings.TrimSpace(value) == "" || strings.EqualFold(value, "true")) && activeServers != "" {
		bootstrap.LogInfo("** Using '%s' servers for active checks", activeServers)
		env["ZBX_ACTIVESERVERS"] = activeServers
	} else {
		delete(env, "ZBX_ACTIVESERVERS")
	}

	delete(env, "ZBX_SERVER_HOST")
	delete(env, "ZBX_SERVER_PORT")

	return SetParameters(configPath,
		Parameter{Name: "Server", Value: env["ZBX_PASSIVESERVERS"]},
		Parameter{Name: "ServerActive", Value: env["ZBX_ACTIVESERVERS"]},
	)
}

// ConfigureAllowDenyKeys writes ZBX_DENYKEY and ZBX_ALLOWKEY into configPath.
func ConfigureAllowDenyKeys(env bootstrap.Environment, configPath string) error {
	if err := MergeParameterValues(configPath, "DenyKey", env["ZBX_DENYKEY"]); err != nil {
		return err
	}

	return MergeParameterValues(configPath, "AllowKey", env["ZBX_ALLOWKEY"])
}

// ProcessTLSFiles persists the agent TLS material from the environment and
// writes the resulting file paths to configPath.
func ProcessTLSFiles(env bootstrap.Environment, homeDir, configPath string) error {
	if err := bootstrap.ProcessTLSFiles(
		env,
		homeDir,
		"ZBX_TLSCA",
		"ZBX_TLSCRL",
		"ZBX_TLSCERT",
		"ZBX_TLSKEY",
		"ZBX_TLSPSK",
	); err != nil {
		return err
	}

	return SetParameters(configPath,
		Parameter{Name: "TLSCAFile", Value: env["ZBX_TLSCAFILE"]},
		Parameter{Name: "TLSCRLFile", Value: env["ZBX_TLSCRLFILE"]},
		Parameter{Name: "TLSCertFile", Value: env["ZBX_TLSCERTFILE"]},
		Parameter{Name: "TLSKeyFile", Value: env["ZBX_TLSKEYFILE"]},
		Parameter{Name: "TLSPSKFile", Value: env["ZBX_TLSPSKFILE"]},
	)
}

// ClearPrivateEnv drops processed ZBX_* variables before the agent starts.
func ClearPrivateEnv(env bootstrap.Environment) {
	bootstrap.ClearPrivateEnv(env, "ZBX_")
}

func prependServer(server, servers string) string {
	if servers == "" {
		return server
	}
	return server + "," + servers
}
