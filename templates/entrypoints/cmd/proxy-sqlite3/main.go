package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
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

	homeDir, err := bootstrap.RequiredHomeDir(env)
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

	if err := sqlite.Prepare(env["ZBX_DB_NAME"], schemaPath); err != nil {
		return err
	}

	if err := proxy.Prepare(env, defaultProxyName); err != nil {
		return err
	}

	return nil
}

func main() {
	bootstrap.Main(bootstrap.Service(proxyBinary, prepareService))
}
