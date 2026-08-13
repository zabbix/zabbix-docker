package main

import (
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/javagateway"
)

func main() {
	bootstrap.Main(javagateway.Run)
}
