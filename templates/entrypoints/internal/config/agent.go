//go:build windows

// Package config prepares Zabbix configuration from the container environment.
package config

import (
	"path/filepath"
	"strings"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

// ConfigureServers merges ZBX_SERVER_HOST and ZBX_SERVER_PORT into the
// passive and active server lists, honouring the ZBX_PASSIVE_ALLOW and
// ZBX_ACTIVE_ALLOW switches.
func ConfigureServers(env bootstrap.Environment) {
	serverHost := env.ValueOrDefault("ZBX_SERVER_HOST", "zabbix-server")
	passiveServers := env["ZBX_PASSIVESERVERS"]
	activeServers := env["ZBX_ACTIVESERVERS"]

	activeServer := serverHost
	if serverPort := env["ZBX_SERVER_PORT"]; serverPort != "" && serverPort != "10051" {
		activeServer += ":" + serverPort
	}
	if serverHost != "" {
		passiveServers = prependServer(serverHost, passiveServers)
		activeServers = prependServer(activeServer, activeServers)
	}

	if value := env["ZBX_PASSIVE_ALLOW"]; (value == "" || strings.EqualFold(value, "true")) && passiveServers != "" {
		bootstrap.LogInfo("** Using '%s' servers for passive checks", passiveServers)
		env["ZBX_PASSIVESERVERS"] = passiveServers
	} else {
		delete(env, "ZBX_PASSIVESERVERS")
	}

	if value := env["ZBX_ACTIVE_ALLOW"]; (value == "" || strings.EqualFold(value, "true")) && activeServers != "" {
		bootstrap.LogInfo("** Using '%s' servers for active checks", activeServers)
		env["ZBX_ACTIVESERVERS"] = activeServers
	} else {
		delete(env, "ZBX_ACTIVESERVERS")
	}

	delete(env, "ZBX_SERVER_HOST")
	delete(env, "ZBX_SERVER_PORT")
}

// ConfigureAllowDenyKeys writes ZBX_DENYKEY and ZBX_ALLOWKEY into the item key
// configuration file.
func ConfigureAllowDenyKeys(env bootstrap.Environment, configDir, fileName string) error {
	configPath := filepath.Join(configDir, fileName)

	if err := MergeParameterValues(configPath, "DenyKey", env["ZBX_DENYKEY"]); err != nil {
		return err
	}

	return MergeParameterValues(configPath, "AllowKey", env["ZBX_ALLOWKEY"])
}

// ProcessTLSFiles persists the agent TLS material from the
// environment into files.
func ProcessTLSFiles(env bootstrap.Environment, homeDir string) error {
	return bootstrap.ProcessTLSFiles(
		env,
		homeDir,
		"ZBX_TLSCA",
		"ZBX_TLSCRL",
		"ZBX_TLSCERT",
		"ZBX_TLSKEY",
		"ZBX_TLSPSK",
	)
}

// ClearPrivateEnv drops internal ZABBIX_* variables before the
// agent starts.
func ClearPrivateEnv(env bootstrap.Environment) {
	bootstrap.ClearPrivateEnv(env, "ZABBIX_")
}

func prependServer(server, servers string) string {
	if servers == "" {
		return server
	}
	return server + "," + servers
}
