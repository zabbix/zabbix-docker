package main

import (
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/hooks"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/postgresql"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/server"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/vault"
)

const (
	serverBinary          = "/usr/sbin/zabbix_server"
	schemaPath            = "/usr/share/doc/zabbix-server-postgresql/create.sql.gz"
	timescaleDBSchemaPath = "/usr/share/doc/zabbix-server-postgresql/timescaledb.sql"
	dbName                = "zabbix"
)

func prepareDatabase(env bootstrap.Environment, db *postgresql.Database) error {
	bootstrap.LogInfo("** Preparing database")

	creds, err := vault.ResolveDatabaseCredentials(env)
	if err != nil {
		return err
	}

	if err := db.Configure(dbName, creds); err != nil {
		return err
	}

	return db.Prepare(schemaPath, timescaleDBSchemaPath)
}

func prepareService(env bootstrap.Environment, db *postgresql.Database) error {
	bootstrap.LogInfo("** Preparing Zabbix server")

	if err := prepareDatabase(env, db); err != nil {
		return err
	}

	db.ApplyRuntimeEnvironment()

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
	bootstrap.ExitOnError(bootstrap.RunDatabaseService(serverBinary,
		func(env bootstrap.Environment) error { return prepareService(env, postgresql.New(env)) },
		func(env bootstrap.Environment) error { return prepareDatabase(env, postgresql.New(env)) },
	))
}
