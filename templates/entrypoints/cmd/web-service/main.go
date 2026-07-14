package main

import (
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/hooks"
)

const webServiceBinary = "/usr/sbin/zabbix_web_service"

func prepareService(env bootstrap.Environment) error {
	bootstrap.LogInfo("** Preparing Zabbix web service")

	homeDirectory, err := bootstrap.RequiredHomeDirectory(env)
	if err != nil {
		return err
	}

	env["ZBX_ALLOWEDIP"] = env.ValueOrDefaultNonEmpty("ZBX_ALLOWEDIP", "zabbix-server")

	if err := bootstrap.ProcessEncryptionFiles(env, homeDirectory, "ZBX_TLSCA", "ZBX_TLSCERT", "ZBX_TLSKEY"); err != nil {
		return err
	}

	if err := hooks.Run(env); err != nil {
		return err
	}

	bootstrap.ClearPrivateEnvironment(env)

	return nil
}

func main() {
	bootstrap.ExitOnError(bootstrap.RunService(webServiceBinary, prepareService))
}
