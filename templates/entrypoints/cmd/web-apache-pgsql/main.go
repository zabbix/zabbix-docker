package main

import (
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/postgresql"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/web"
)

func main() {
	entry := web.Entrypoint(postgresql.NewForFrontend, web.Options{DBType: web.PostgreSQL, Server: web.Apache})
	bootstrap.Main(bootstrap.WithResolvedSecrets(bootstrap.SecretsWebPgSQL, "", entry))
}
