package main

import (
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/snmptraps"
)

func main() {
	bootstrap.Main(snmptraps.Run)
}
