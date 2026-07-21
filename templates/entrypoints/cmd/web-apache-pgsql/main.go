package main

import (
	"os"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/postgresql"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/web"
)

func main() {
	env := bootstrap.NewEnvironment(os.Environ())

	if len(os.Args) == 1 {
		bootstrap.LogInfo("** Preparing Zabbix web-interface (Apache) with PostgreSQL database")
	}

	err := web.Run(env, postgresql.NewForFrontend(env), web.Options{DBType: web.PostgreSQL, Server: web.Apache}, os.Args[1:])

	bootstrap.ExitOnError(err)
}
