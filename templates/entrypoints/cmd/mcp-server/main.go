package main

import (
	"path/filepath"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/config"
)

const mcpServerBinary = "/usr/sbin/zabbix_mcp_server"

func prepareService(env bootstrap.Environment, runtimeRulePath string) error {
	bootstrap.LogInfo("** Preparing Zabbix MCP server")

	homeDir, configDir, err := bootstrap.CommonDirs(env)
	if err != nil {
		return err
	}

	env["ZBX_ALLOWEDIP"] = env.ValueOrDefaultNonEmpty("ZBX_ALLOWEDIP", "127.0.0.1,::1")

	if err := config.ConfigureToolAccessRules(env,
		filepath.Join(configDir, "zabbix_mcp_server_tool_access.conf"), runtimeRulePath,
	); err != nil {
		return err
	}

	if err := bootstrap.ProcessTLSFiles(env, filepath.Join(homeDir, "enc"), "ZBX_TLSCERT", "ZBX_TLSKEY"); err != nil {
		return err
	}

	bootstrap.ClearPrivateEnv(env)

	return nil
}

func main() {
	bootstrap.Main(bootstrap.Service(mcpServerBinary, func(env bootstrap.Environment) error {
		return prepareService(env, "/tmp/zabbix_mcp_server_tool_access.conf")
	}))
}
