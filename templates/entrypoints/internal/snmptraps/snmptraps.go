// Package snmptraps implements the Zabbix SNMP traps container entrypoint.
package snmptraps

import (
	"path/filepath"
	"strings"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

const (
	snmptrapdBinary = "/usr/sbin/snmptrapd"
	mainConfigFile  = "/etc/snmp/snmptrapd.conf"

	// SNMP trap output options:
	// S - display the MIB name as well as the object name;
	// T - display a printable version of hexadecimal strings;
	// t - display TimeTicks values as raw numbers;
	// e - remove symbolic labels from enumeration values.
	defaultOutputOptions = "STte"
)

// Run prepares snmptrapd when the image default command or snmptrapd options
// are requested. Other commands are executed unchanged.
func Run(env bootstrap.Environment, args []string) error {
	extraArgs, startsService := bootstrap.ServiceArgs(args, snmptrapdBinary)
	if !startsService {
		return bootstrap.Execute(args, env)
	}

	env.SetDefaultNonEmpty("SNMPTRAP_OUTPUT_OPTIONS", defaultOutputOptions)

	command := buildCommand(env, extraArgs)

	return bootstrap.Execute(command, env)
}

// buildCommand constructs the default snmptrapd pipeline. Persistent
// configuration files are loaded after the image defaults, allowing an
// explicitly supplied user configuration to override them.
func buildCommand(env bootstrap.Environment, extraArgs []string) []string {
	configFiles := []string{mainConfigFile}

	if persistentDir := env["SNMP_PERSISTENT_DIR"]; persistentDir != "" {
		for _, name := range []string{"snmptrapd.conf", "snmptrapd_custom.conf"} {
			path := filepath.Join(persistentDir, name)

			if bootstrap.RegularFile(path) {
				configFiles = append(configFiles, path)
			}
		}
	}

	command := []string{
		snmptrapdBinary,
		"-f",
		"-a",
		"-C", "-c", strings.Join(configFiles, ","),
		"-t",
		"-X",
		"-Lo",
		"--hexOutputLength=0",
		"-O" + env["SNMPTRAP_OUTPUT_OPTIONS"],
	}

	if env["ZBX_SNMP_TRAP_USE_DNS"] != "true" {
		command = append(command, "-n")
	}

	if env["DEBUG_MODE"] == "true" {
		command = append(command, "-DALL")
	}

	command = append(command, extraArgs...)

	return append(command, "udp:1162", "udp6:1162")
}
