package main

import "github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"

const webServiceBinary = "/usr/sbin/zabbix_web_service"

func prepareService(env bootstrap.Environment) error {
	bootstrap.LogInfo("** Preparing Zabbix web service")

	env.SetDefaultNonEmpty("ZBX_ALLOWEDIP", "zabbix-server")

	bootstrap.ClearPrivateEnv(env)

	return nil
}

func main() {
	entry := bootstrap.Service(webServiceBinary, prepareService)
	bootstrap.Main(bootstrap.WithResolvedSecrets(bootstrap.SecretsWebService, webServiceBinary, entry))
}
