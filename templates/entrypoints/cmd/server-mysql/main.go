package main

import (
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/mysql"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/server"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/storageinit"
)

const (
	serverBinary = "/usr/sbin/zabbix_server"
	schemaPath   = "/usr/share/doc/zabbix-server-mysql/mysql/create.sql.gz"
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
	return server.Prepare(env)
}

func main() {
	entry := bootstrap.DBService(serverBinary,
		func(env bootstrap.Environment) error { return prepareService(env, mysql.NewForBackend(env)) },
		func(env bootstrap.Environment) error { return prepareDB(env, mysql.NewForBackend(env)) },
	)
	bootstrap.Main(bootstrap.StorageInitEntrypoint(entry, storageinit.PrepareHistory, storageinit.PrepareTelemetry))
}
