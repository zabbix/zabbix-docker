// Package proxy prepares the runtime environment for Zabbix proxy.
package proxy

import (
	"path/filepath"
	"strings"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

// Prepare performs the proxy-specific entrypoint steps. defaultHostname is
// used when neither ZBX_HOSTNAME nor ZBX_HOSTNAMEITEM is set.
func Prepare(env bootstrap.Environment, defaultHostname string) error {
	env["ZBX_SERVER_HOST"] = env.ValueOrDefaultNonEmpty("ZBX_SERVER_HOST", "zabbix-server")

	// Keep the hostname explicitly empty when ZBX_HOSTNAMEITEM is used:
	// the proxy then resolves its own name via the item.
	hostname := ""
	if env["ZBX_HOSTNAME"] != "" || env["ZBX_HOSTNAMEITEM"] == "" {
		hostname = env.ValueOrDefaultNonEmpty("ZBX_HOSTNAME", defaultHostname)
	}
	env["ZBX_HOSTNAME"] = hostname

	if strings.EqualFold(env.ValueOrDefaultNonEmpty("ZBX_ENABLE_SNMP_TRAPS", "false"), "true") {
		env["ZBX_STARTSNMPTRAPPER"] = "1"
	}
	delete(env, "ZBX_ENABLE_SNMP_TRAPS")

	homeDir, configDir, err := bootstrap.RequiredDirectories(env)
	if err != nil {
		return err
	}

	if err := bootstrap.UpdateConfigMultiple(filepath.Join(configDir, "zabbix_proxy_modules.conf"), "LoadModule", env["ZBX_LOADMODULE"]); err != nil {
		return err
	}

	if err := bootstrap.ProcessTLSFiles(env, homeDir, "ZBX_TLSCA", "ZBX_TLSCRL", "ZBX_TLSCERT", "ZBX_TLSKEY", "ZBX_TLSPSK"); err != nil {
		return err
	}

	if err := bootstrap.ConfigureRunUser(env); err != nil {
		return err
	}

	bootstrap.RehashCertDir(env["ZBX_SSLCALOCATION"])

	return nil
}
