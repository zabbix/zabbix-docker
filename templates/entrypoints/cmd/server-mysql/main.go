package main

import (
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/hooks"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/mysql"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/server"
)

const (
	serverBinary = "/usr/sbin/zabbix_server"
	schemaPath   = "/usr/share/doc/zabbix-server-mysql/create.sql.gz"
	dbName       = "zabbix"
)

func prepareDB(env bootstrap.Environment, db *mysql.DB) error {
	bootstrap.LogInfo("** Preparing database")

	if err := db.Configure(dbName); err != nil {
		return err
	}

	return db.Prepare(schemaPath)
}

func prepareService(env bootstrap.Environment, db *mysql.DB) error {
	bootstrap.LogInfo("** Preparing Zabbix server")

	if err := prepareDB(env, db); err != nil {
		return err
	}

	db.ExportEnv()
	if err := server.Prepare(env); err != nil {
		return err
	}

	if err := hooks.Run(env); err != nil {
		return err
	}

	bootstrap.ClearPrivateEnv(env)

	return nil
}

func main() {
	bootstrap.ExitOnError(bootstrap.RunDBService(serverBinary,
		func(env bootstrap.Environment) error { return prepareService(env, mysql.NewForBackend(env)) },
		func(env bootstrap.Environment) error { return prepareDB(env, mysql.NewForBackend(env)) },
	))
}
