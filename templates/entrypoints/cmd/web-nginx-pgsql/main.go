package main

import (
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/postgresql"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/web"
)

func main() {
	bootstrap.Main(web.Entrypoint(postgresql.NewForFrontend, web.Options{DBType: web.PostgreSQL, Server: web.Nginx}))
}
