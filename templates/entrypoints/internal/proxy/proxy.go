// Package proxy prepares the runtime environment for Zabbix proxy.
package proxy

import (
	"path/filepath"
	"strings"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/config"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/hooks"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/storageinit"
)

// Prepare performs the proxy-specific entrypoint steps. defaultHostname is
// used when neither ZBX_HOSTNAME nor ZBX_HOSTNAMEITEM is set.
func Prepare(env bootstrap.Environment, defaultHostname string) error {
	env.SetDefaultNonEmpty("ZBX_SERVER_HOST", "zabbix-server")

	// Keep the hostname explicitly empty when ZBX_HOSTNAMEITEM is used:
	// the proxy then resolves its own name via the item.
	hostname := ""
	if env["ZBX_HOSTNAME"] != "" || env["ZBX_HOSTNAMEITEM"] == "" {
		hostname = env.ValueOrDefaultNonEmpty("ZBX_HOSTNAME", defaultHostname)
	}
	env["ZBX_HOSTNAME"] = hostname

	if strings.EqualFold(env["ZBX_ENABLE_SNMP_TRAPS"], "true") {
		env["ZBX_STARTSNMPTRAPPER"] = "1"
	}
	delete(env, "ZBX_ENABLE_SNMP_TRAPS")

	_, configDir, err := bootstrap.CommonDirs(env)
	if err != nil {
		return err
	}

	if err := config.MergeParameterValues(filepath.Join(configDir, "zabbix_proxy_modules.conf"), "LoadModule", env["ZBX_LOADMODULE"]); err != nil {
		return err
	}
	if err := storageinit.ConfigureTelemetry(env); err != nil {
		return err
	}
	if err := config.UpdateIndexedParameter(env, filepath.Join(configDir, "zabbix_proxy_telemetry.conf"), "TelemetryProvider", "ZBX_TELEMETRYPROVIDER"); err != nil {
		return err
	}

	bootstrap.ConfigureRunUser(env)

	if err := hooks.Run(env); err != nil {
		return err
	}

	bootstrap.ClearPrivateEnv(env)

	return nil
}
