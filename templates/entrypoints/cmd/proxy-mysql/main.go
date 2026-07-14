package main

import (
	"os"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/hooks"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/mysql"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/proxy"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/vault"
)

const (
	proxyBinary      = "/usr/sbin/zabbix_proxy"
	schemaPath       = "/usr/share/doc/zabbix-proxy-mysql/create.sql.gz"
	dbName           = "zabbix_proxy"
	defaultProxyName = "zabbix-proxy-mysql"
)

func prepareDatabase(env bootstrap.Environment, db *mysql.Database) error {
	bootstrap.LogInfo("** Preparing database")

	creds, err := vault.ResolveDatabaseCredentials(env)
	if err != nil {
		return err
	}

	if err := db.Configure(dbName, creds); err != nil {
		return err
	}

	return db.Prepare(schemaPath)
}

func prepareService(env bootstrap.Environment, db *mysql.Database) error {
	bootstrap.LogInfo("** Preparing Zabbix proxy")

	if err := prepareDatabase(env, db); err != nil {
		return err
	}
	db.ApplyRuntimeEnvironment()

	if err := proxy.Prepare(env, defaultProxyName); err != nil {
		return err
	}

	if err := hooks.Run(env); err != nil {
		return err
	}

	bootstrap.ClearPrivateEnv(env)

	return nil
}

func run() error {
	env := bootstrap.NewEnvironment(os.Environ())

	db := mysql.New(env)

	args := bootstrap.Command(os.Args[1:], proxyBinary)
	if args[0] == proxyBinary {
		if err := prepareService(env, db); err != nil {
			return err
		}
	}

	if args[0] == "init_db_only" {
		return prepareDatabase(env, db)
	}

	return bootstrap.Execute(args, env)
}

func main() {
	bootstrap.ExitOnError(run())
}
