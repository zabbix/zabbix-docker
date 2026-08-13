package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/hooks"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/proxy"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/sqlite"
)

const (
	proxyBinary      = "/usr/sbin/zabbix_proxy"
	schemaPath       = "/usr/share/doc/zabbix-proxy-sqlite3/create.sql.gz"
	defaultProxyName = "zabbix-proxy-sqlite3"
)

func prepareService(env bootstrap.Environment) error {
	bootstrap.LogInfo("** Preparing Zabbix proxy")

	homeDir, err := bootstrap.RequiredHomeDirectory(env)
	if err != nil {
		return err
	}

	proxyName := env.ValueOrDefaultNonEmpty("ZBX_HOSTNAME", defaultProxyName)

	if strings.EqualFold(env["ZBX_USE_NODE_NAME_AS_DB_NAME"], "true") {
		proxyName, err = os.Hostname()
		if err != nil {
			return err
		}
	}
	delete(env, "ZBX_USE_NODE_NAME_AS_DB_NAME")

	env["ZBX_DB_NAME"] = filepath.Join(homeDir, "db_data", proxyName+".sqlite")

	if err := proxy.Prepare(env, proxyName); err != nil {
		return err
	}

	if err := sqlite.Prepare(env["ZBX_DB_NAME"], schemaPath); err != nil {
		return err
	}

	if err := hooks.Run(env); err != nil {
		return err
	}

	bootstrap.ClearPrivateEnv(env)

	return nil
}

func main() {
	bootstrap.Main(bootstrap.Service(proxyBinary, prepareService))
}
