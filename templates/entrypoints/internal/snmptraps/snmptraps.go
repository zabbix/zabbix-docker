// Package snmptraps implements the Zabbix SNMP traps container entrypoint.
package snmptraps

import (
	"path/filepath"
	"strings"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/hooks"
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

	if err := hooks.Run(env); err != nil {
		return err
	}

	command := buildCommand(env, extraArgs)

	return bootstrap.Execute(command, env)
}

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
		"--doNotFork=yes",
		"-C", "-c", strings.Join(configFiles, ","),
	}

	if env["ZBX_SNMP_TRAP_USE_DNS"] != "true" {
		command = append(command, "-n")
	}

	command = append(command,
		"-t",
		"-X", "-Lo", "-A",
		"-O"+env["SNMPTRAP_OUTPUT_OPTIONS"],
	)

	return append(command, extraArgs...)
}
