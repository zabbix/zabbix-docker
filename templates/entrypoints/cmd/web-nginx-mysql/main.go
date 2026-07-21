package main

import (
	"os"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/mysql"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/web"
)

func main() {
	env := bootstrap.NewEnvironment(os.Environ())

	if len(os.Args) == 1 {
		bootstrap.LogInfo("** Preparing Zabbix web-interface (Nginx) with MySQL database")
	}

	bootstrap.ExitOnError(web.Run(env, mysql.NewForFrontend(env), web.Options{DBType: web.MySQL, Server: web.Nginx}, os.Args[1:]))
}
