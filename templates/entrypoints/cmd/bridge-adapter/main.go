package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/config"
)

const (
	bridgeAdapterBinary      = "/usr/sbin/zabbix_bridge_adapter"
	allowedMethodsConfigFile = "zabbix_bridge_adapter_allowed_methods.conf"
	acceptLicenseCommand     = "accept-license"
)

var runtimeAllowedMethodsConfigPath = "/tmp/" + allowedMethodsConfigFile

func prepareLicenseAcceptance(env bootstrap.Environment) ([]string, error) {
	version := strings.TrimSpace(env["ZBX_EULAVERSION"])
	if version == "" {
		return nil, fmt.Errorf("accept-license: set ZBX_EULAVERSION to the reviewed bundled EULA version; no license has been accepted")
	}

	configDir := env.ValueOrDefaultNonEmpty("ZABBIX_CONF_DIR", "/etc/zabbix")
	command := []string{
		bridgeAdapterBinary, "-c", filepath.Join(configDir, "zabbix_bridge_adapter.conf"),
		"license", "accept", "--non-interactive", "--version", version,
	}
	bootstrap.LogInfo("** Requesting non-interactive EULA acceptance for version %q supplied by ZBX_EULAVERSION", version)
	bootstrap.LogInfo("** The Adapter will verify the bundled EULA version and write acceptance evidence to its configured DataDir")
	bootstrap.LogInfo("** Configuration files will not be rewritten; Adapter service will not be started")
	bootstrap.LogInfo("** Executing: %q", command)

	return command, nil
}

func prepareService(env bootstrap.Environment) error {
	bootstrap.LogInfo("** Preparing Zabbix Bridge Adapter")

	homeDir, err := bootstrap.RequiredHomeDir(env)
	if err != nil {
		return err
	}
	if err := bootstrap.ProcessTLSFiles(
		env,
		filepath.Join(homeDir, "enc"),
		"ZBX_TLSCA", "ZBX_TLSCERT", "ZBX_TLSKEY", "ZBX_ZABBIXAPITLSCA",
	); err != nil {
		return err
	}

	variables, err := bootstrap.CollectIndexed(env, []string{"ZBX_ALLOWEDAPIMETHODS"}, false)
	if err != nil {
		return err
	}
	if len(variables) == 0 {
		bootstrap.ClearPrivateEnv(env)
		return nil
	}

	configDir, err := bootstrap.RequiredConfigDir(env)
	if err != nil {
		return err
	}

	templatePath := filepath.Join(configDir, allowedMethodsConfigFile)
	data, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("read security configuration template %s: %w", templatePath, err)
	}

	if err := os.WriteFile(runtimeAllowedMethodsConfigPath, data, 0o600); err != nil {
		return fmt.Errorf("write runtime allowed methods configuration %s: %w", runtimeAllowedMethodsConfigPath, err)
	}

	if err := config.UpdateIndexedParameter(
		env,
		runtimeAllowedMethodsConfigPath,
		"AllowedAPIMethods",
		"ZBX_ALLOWEDAPIMETHODS",
	); err != nil {
		return err
	}

	bootstrap.ClearPrivateEnv(env)

	return nil
}

func prepareCommand(env bootstrap.Environment, args []string) ([]string, error) {
	if len(args) > 0 && args[0] == acceptLicenseCommand {
		if len(args) != 1 {
			return nil, fmt.Errorf("usage: accept-license (set ZBX_EULAVERSION in the environment)")
		}
		command, err := prepareLicenseAcceptance(env)
		if err != nil {
			return nil, err
		}
		bootstrap.ClearPrivateEnv(env)
		return command, nil
	}

	serviceArgs, startsService := bootstrap.ServiceArgs(args, bridgeAdapterBinary)
	if !startsService {
		return args, nil
	}

	command := append([]string{bridgeAdapterBinary}, serviceArgs...)
	if err := prepareService(env); err != nil {
		return nil, err
	}
	return command, nil
}

func entrypoint(env bootstrap.Environment, args []string) error {
	command, err := prepareCommand(env, args)
	if err != nil {
		return err
	}
	return bootstrap.Execute(command, env)
}

func main() {
	bootstrap.Main(entrypoint)
}
