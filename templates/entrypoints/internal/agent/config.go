package agent

import (
	"path/filepath"
	"strings"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

func ConfigureServers(env bootstrap.Environment) {
	serverHost := env.ValueOrDefault("ZBX_SERVER_HOST", "zabbix-server")
	serverPort := env.ValueOrDefault("ZBX_SERVER_PORT", "10051")
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

	if allowed(env["ZBX_PASSIVE_ALLOW"]) && passiveServers != "" {
		bootstrap.LogInfo("** Using '%s' servers for passive checks", passiveServers)
		env["ZBX_PASSIVESERVERS"] = passiveServers
	} else {
		delete(env, "ZBX_PASSIVESERVERS")
	}

	if allowed(env["ZBX_ACTIVE_ALLOW"]) && activeServers != "" {
		bootstrap.LogInfo("** Using '%s' servers for active checks", activeServers)
		env["ZBX_ACTIVESERVERS"] = activeServers
	} else {
		delete(env, "ZBX_ACTIVESERVERS")
	}

	delete(env, "ZBX_SERVER_HOST")
	delete(env, "ZBX_SERVER_PORT")
}

func ConfigureItemKeys(env bootstrap.Environment, configDirectory, fileName string) error {
	path := filepath.Join(configDirectory, fileName)

	if err := bootstrap.UpdateConfigMultiple(path, "DenyKey", env["ZBX_DENYKEY"]); err != nil {
		return err
	}

	return bootstrap.UpdateConfigMultiple(path, "AllowKey", env["ZBX_ALLOWKEY"])
}

func ProcessEncryptionFiles(env bootstrap.Environment, homeDirectory string) error {
	return bootstrap.ProcessEncryptionFiles(
		env,
		homeDirectory,
		"ZBX_TLSCA",
		"ZBX_TLSCRL",
		"ZBX_TLSCERT",
		"ZBX_TLSKEY",
		"ZBX_TLSPSK",
	)
}

func ClearPrivateEnvironment(env bootstrap.Environment) {
	bootstrap.ClearPrivateEnvironment(env, "ZABBIX_")
}

func prependServer(server, servers string) string {
	if servers == "" {
		return server
	}
	return server + "," + servers
}

func allowed(value string) bool {
	return value == "" || strings.EqualFold(value, "true")
}
