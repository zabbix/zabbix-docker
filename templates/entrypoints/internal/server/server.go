// Package server prepares the runtime environment for Zabbix server.
package server

import (
	"path/filepath"
	"strings"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

// Prepare performs the server-specific entrypoint steps: modules, history
// storage providers, TLS files, HA node autoconfiguration and CA
// certificate rehashing.
func Prepare(env bootstrap.Environment) error {
	if strings.EqualFold(env.ValueOrDefaultNonEmpty("ZBX_ENABLE_SNMP_TRAPS", "false"), "true") {
		env["ZBX_STARTSNMPTRAPPER"] = "1"
	}
	delete(env, "ZBX_ENABLE_SNMP_TRAPS")

	homeDir, configDir, err := bootstrap.RequiredDirectories(env)
	if err != nil {
		return err
	}

	if err := bootstrap.UpdateConfigMultiple(filepath.Join(configDir, "zabbix_server_modules.conf"), "LoadModule", env["ZBX_LOADMODULE"]); err != nil {
		return err
	}

	if err := bootstrap.UpdateConfigIndexed(env, filepath.Join(configDir, "zabbix_server_history_storage.conf"), "HistoryProvider", "ZBX_HISTORYPROVIDER"); err != nil {
		return err
	}

	if err := bootstrap.ProcessTLSFiles(env, homeDir, "ZBX_TLSCA", "ZBX_TLSCRL", "ZBX_TLSCERT", "ZBX_TLSKEY"); err != nil {
		return err
	}

	if err := configureHANode(env); err != nil {
		return err
	}

	if err := bootstrap.ConfigureRunUser(env); err != nil {
		return err
	}

	bootstrap.RehashCertificateDirectory(env["ZBX_SSLCALOCATION"])

	return nil
}

// configureHANode fills ZBX_HANODENAME and ZBX_NODEADDRESS from the host
// name when the corresponding ZBX_AUTO* variable requests it.
func configureHANode(env bootstrap.Environment) error {
	if env["ZBX_HANODENAME"] == "" {
		hostname, err := autoHostname(env["ZBX_AUTOHANODENAME"])
		if err != nil {
			return err
		}
		if hostname != "" {
			env["ZBX_HANODENAME"] = hostname
		}
	}
	delete(env, "ZBX_AUTOHANODENAME")

	if env["ZBX_NODEADDRESS"] == "" {
		hostname, err := autoHostname(env["ZBX_AUTONODEADDRESS"])
		if err != nil {
			return err
		}
		if hostname != "" {
			env["ZBX_NODEADDRESS"] = hostname + ":" + env.ValueOrDefaultNonEmpty("ZBX_NODEADDRESSPORT", "10051")
		}
	}
	delete(env, "ZBX_AUTONODEADDRESS")

	return nil
}

func autoHostname(mode string) (string, error) {
	if mode != "fqdn" && mode != "hostname" {
		return "", nil
	}
	return bootstrap.Hostname(mode == "fqdn")
}
