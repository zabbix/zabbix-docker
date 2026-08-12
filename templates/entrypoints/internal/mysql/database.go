// Package mysql provisions the MySQL database for Zabbix server, proxy and
// the web frontend images.
package mysql

import (
	"cmp"
	"fmt"
	"net"
	"strings"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/vault"
)

const zabbixDBPrivileges = "SELECT, INSERT, UPDATE, DELETE, CREATE, DROP, ALTER, INDEX, CREATE TEMPORARY TABLES, TRIGGER, REFERENCES"

// DB carries the resolved MySQL connection target and the working and
// administrative credentials.
type DB struct {
	env        bootstrap.Environment
	tls        bootstrap.DBTLSConfig
	open       sessionOpener
	network    string
	address    string
	charset    string
	collation  string
	adminUser  string
	adminPass  string
	user       string
	password   string
	name       string
	createUser bool
	fromVault  bool
}

// NewForBackend creates an unconfigured DB for a Zabbix backend service;
// call Configure before use.
func NewForBackend(env bootstrap.Environment) *DB {
	return newDB(env, bootstrap.ServiceDBTLS(env))
}

// NewForFrontend creates a DB whose bootstrap connection uses the
// PHP frontend's ZBX_DB_* TLS settings.
func NewForFrontend(env bootstrap.Environment) *DB {
	return newDB(env, bootstrap.FrontendDBTLS(env))
}

func newDB(env bootstrap.Environment, tls bootstrap.DBTLSConfig) *DB {
	return &DB{
		env:       env,
		tls:       tls,
		open:      openDBSession,
		charset:   env.ValueOrDefaultNonEmpty("DB_CHARACTER_SET", "utf8mb4"),
		collation: env.ValueOrDefaultNonEmpty("DB_CHARACTER_COLLATE", "utf8mb4_bin"),
	}
}

// Configure resolves the connection target and credentials from the
// environment or Vault.
func (db *DB) Configure(defaultDBName string) error {
	if socket := db.env["DB_SERVER_SOCKET"]; socket != "" {
		db.network = "unix"
		db.address = socket
	} else {
		host := db.env.ValueOrDefaultNonEmpty("DB_SERVER_HOST", "mysql-server")
		port, err := bootstrap.ResolveDBPort(db.env, "3306")
		if err != nil {
			return err
		}
		db.env["DB_SERVER_HOST"] = host
		db.network = "tcp"
		db.address = net.JoinHostPort(strings.Trim(host, "[]"), port)
	}

	creds, err := vault.ResolveDBCredentials(db.env)
	if err != nil {
		return err
	}

	credVars := []string{"MYSQL_ROOT_USER", "MYSQL_ROOT_PASSWORD"}
	if creds == nil {
		credVars = append(credVars, "MYSQL_USER", "MYSQL_PASSWORD")
	}

	for _, name := range credVars {
		if err := bootstrap.ResolveSecretEnv(db.env, name); err != nil {
			return err
		}
	}

	mysqlUser := db.env["MYSQL_USER"]
	mysqlPassword := db.env["MYSQL_PASSWORD"]
	if creds != nil {
		mysqlUser = creds.Username
		mysqlPassword = creds.Password
	}

	mysqlRootUser := db.env["MYSQL_ROOT_USER"]
	mysqlRootPassword := db.env["MYSQL_ROOT_PASSWORD"]
	allowEmptyPassword := db.env["MYSQL_ALLOW_EMPTY_PASSWORD"] == "true"

	if mysqlUser == "" && mysqlRootPassword == "" && !allowEmptyPassword {
		return fmt.Errorf("impossible to use MySQL server because MYSQL_ROOT_PASSWORD is not defined and empty password is not allowed")
	}

	useMySQLRoot := allowEmptyPassword || mysqlRootPassword != ""
	if useMySQLRoot {
		db.adminUser = cmp.Or(mysqlRootUser, "root")
		db.adminPass = mysqlRootPassword
	} else {
		db.adminUser = db.env["DB_SERVER_ROOT_USER"]
		db.adminPass = db.env["DB_SERVER_ROOT_PASS"]
	}

	_, externalAdminPassSet := db.env["DB_SERVER_ROOT_PASS"]
	hasAdminCredentials := useMySQLRoot || db.adminUser != "" && externalAdminPassSet
	db.createUser = mysqlUser != "" && hasAdminCredentials
	if db.adminUser == "" {
		db.adminUser = mysqlUser
		db.adminPass = mysqlPassword
	}

	db.user = cmp.Or(mysqlUser, "zabbix")
	db.password = cmp.Or(mysqlPassword, "zabbix")
	db.name = db.env.ValueOrDefaultNonEmpty("MYSQL_DATABASE", defaultDBName)
	db.fromVault = creds != nil

	return nil
}

// ExportEnv exports the resolved connection settings as
// ZBX_DB_* variables for the service.
func (db *DB) ExportEnv() {
	if db.env["DB_SERVER_SOCKET"] != "" {
		db.env["ZBX_DB_SOCKET"] = db.env["DB_SERVER_SOCKET"]
	}
	if db.env["DB_SERVER_HOST"] != "" {
		db.env["ZBX_DB_HOST"] = db.env["DB_SERVER_HOST"]
	}
	if db.env["DB_SERVER_PORT"] != "" {
		db.env["ZBX_DB_PORT"] = db.env["DB_SERVER_PORT"]
	}
	db.env["ZBX_DB_NAME"] = db.name

	if db.fromVault {
		delete(db.env, "ZBX_DB_USER")
		delete(db.env, "ZBX_DB_PASSWORD")
	} else {
		db.env["ZBX_DB_USER"] = db.user
		db.env["ZBX_DB_PASSWORD"] = db.password
	}
}

func (db *DB) Name() string     { return db.name }
func (db *DB) Schema() string   { return "" }
func (db *DB) User() string     { return db.user }
func (db *DB) Password() string { return db.password }
