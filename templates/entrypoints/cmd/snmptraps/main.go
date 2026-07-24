package main

import (
	"os"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/snmptraps"
)

func main() {
	env := bootstrap.NewEnvironment(os.Environ())
	bootstrap.ExitOnError(snmptraps.Run(env, os.Args[1:]))
}
