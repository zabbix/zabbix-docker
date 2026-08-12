package main

import (
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/mysql"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/web"
)

func main() {
	bootstrap.Main(web.Entrypoint(mysql.NewForFrontend, web.Options{DBType: web.MySQL, Server: web.Nginx}))
}
