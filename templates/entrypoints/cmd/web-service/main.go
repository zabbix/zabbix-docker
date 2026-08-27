package main

import "github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"

const webServiceBinary = "/usr/sbin/zabbix_web_service"

func prepareService(env bootstrap.Environment) error {
	bootstrap.LogInfo("** Preparing Zabbix web service")

	homeDir, err := bootstrap.RequiredHomeDir(env)
	if err != nil {
		return err
	}

	env["ZBX_ALLOWEDIP"] = env.ValueOrDefaultNonEmpty("ZBX_ALLOWEDIP", "zabbix-server")

	if err := bootstrap.ProcessTLSFiles(env, homeDir, "ZBX_TLSCA", "ZBX_TLSCERT", "ZBX_TLSKEY"); err != nil {
		return err
	}

	bootstrap.ClearPrivateEnv(env)

	return nil
}

func main() {
	bootstrap.Main(bootstrap.Service(webServiceBinary, prepareService))
}
