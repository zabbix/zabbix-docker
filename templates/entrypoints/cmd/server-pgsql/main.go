package main

import (
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/postgresql"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/server"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/storageinit"
)

const (
	serverBinary          = "/usr/sbin/zabbix_server"
	schemaPath            = "/usr/share/doc/zabbix-server-postgresql/postgresql/create.sql.gz"
	timescaleDBSchemaPath = "/usr/share/doc/zabbix-server-postgresql/postgresql/timescaledb.sql"
	dbName                = "zabbix"
)

func prepareDB(env bootstrap.Environment, db *postgresql.DB) error {
	bootstrap.LogInfo("** Preparing database")

	if err := db.Configure(dbName); err != nil {
		return err
	}

	return db.Prepare(schemaPath, timescaleDBSchemaPath)
}

func prepareService(env bootstrap.Environment, db *postgresql.DB) error {
	bootstrap.LogInfo("** Preparing Zabbix server")

	if err := prepareDB(env, db); err != nil {
		return err
	}

	db.ExportEnv()

	return server.Prepare(env)
}

func main() {
	entry := bootstrap.DBService(serverBinary,
		func(env bootstrap.Environment) error { return prepareService(env, postgresql.NewForBackend(env)) },
		bootstrap.WithInitDBSecrets(bootstrap.SecretsServerPgSQL,
			func(env bootstrap.Environment) error { return prepareDB(env, postgresql.NewForBackend(env)) }),
	)
	entry = bootstrap.StorageInitEntrypoint(entry,
		bootstrap.WithHistoryInitSecrets(bootstrap.SecretsServerPgSQL, storageinit.PrepareHistory),
		bootstrap.WithTelemetryInitSecrets(bootstrap.SecretsServerPgSQL, storageinit.PrepareTelemetry))
	bootstrap.Main(bootstrap.WithResolvedSecrets(bootstrap.SecretsServerPgSQL, serverBinary, entry))
}
