// Package mysql provisions the MySQL database for Zabbix server, proxy and
// the web frontend images.
package mysql

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

const zabbixDatabasePrivileges = "SELECT, INSERT, UPDATE, DELETE, CREATE, DROP, ALTER, INDEX, CREATE TEMPORARY TABLES, TRIGGER, REFERENCES"

// Database carries the resolved MySQL connection target and the working and
// administrative credentials.
type Database struct {
	env              bootstrap.Environment
	sleep            func(time.Duration)
	open             sessionOpener
	network          string
	address          string
	characterSet     string
	characterCollate string
	rootUser         string
	rootPassword     string
	zabbixUser       string
	zabbixPassword   string
	name             string
	createUser       bool
}

// New creates an unconfigured Database; call Configure before use.
func New(env bootstrap.Environment) *Database {
	return &Database{
		env:              env,
		sleep:            time.Sleep,
		open:             openDatabaseSession,
		characterSet:     env.ValueOrDefaultNonEmpty("DB_CHARACTER_SET", "utf8mb4"),
		characterCollate: env.ValueOrDefaultNonEmpty("DB_CHARACTER_COLLATE", "utf8mb4_bin"),
	}
}

// Configure resolves the connection target and the credentials from the
// environment, following the same rules as the shell entrypoints did:
// MYSQL_* variables with the *_FILE secret convention, optional root
// access and credentials coming from Vault.
func (db *Database) Configure(defaultDBName string, creds *bootstrap.DBCredentials) error {
	if socket := db.env["DB_SERVER_SOCKET"]; socket != "" {
		db.network = "unix"
		db.address = socket
	} else {
		host := db.env.ValueOrDefaultNonEmpty("DB_SERVER_HOST", "mysql-server")
		port := db.env.ValueOrDefaultNonEmpty("DB_SERVER_PORT", "3306")
		db.env["DB_SERVER_HOST"] = host
		db.env["DB_SERVER_PORT"] = port
		db.network = "tcp"
		db.address = net.JoinHostPort(strings.Trim(host, "[]"), port)
	}

	credVars := []string{"MYSQL_ROOT_USER", "MYSQL_ROOT_PASSWORD"}
	if creds == nil {
		credVars = append(credVars, "MYSQL_USER", "MYSQL_PASSWORD")
	}
	for _, name := range credVars {
		if err := bootstrap.FileEnv(db.env, name, ""); err != nil {
			return err
		}
	}

	mysqlUser := db.env["MYSQL_USER"]
	mysqlPassword := db.env["MYSQL_PASSWORD"]
	if creds != nil {
		mysqlUser = creds.Username
		mysqlPassword = creds.Password
	}

	mysqlRootPassword := db.env["MYSQL_ROOT_PASSWORD"]
	allowEmptyPassword := db.env["MYSQL_ALLOW_EMPTY_PASSWORD"] == "true"

	if mysqlUser == "" && db.env["MYSQL_RANDOM_ROOT_PASSWORD"] == "true" {
		return fmt.Errorf("impossible to use MySQL server because of unknown Zabbix user and random 'root' password")
	}
	if mysqlUser == "" && mysqlRootPassword == "" && !allowEmptyPassword {
		return fmt.Errorf("impossible to use MySQL server because 'root' password is not defined and empty password is not allowed")
	}

	useRootUser := allowEmptyPassword || mysqlRootPassword != ""
	if useRootUser {
		db.rootUser = db.env.ValueOrDefaultNonEmpty("MYSQL_ROOT_USER", "root")
		db.rootPassword = mysqlRootPassword
	} else {
		db.rootUser = db.env["DB_SERVER_ROOT_USER"]
		db.rootPassword = db.env["DB_SERVER_ROOT_PASS"]
	}

	db.createUser = creds == nil && mysqlUser != "" && useRootUser
	if db.rootUser == "" {
		db.rootUser = mysqlUser
	}
	if !allowEmptyPassword && db.rootPassword == "" {
		db.rootPassword = mysqlPassword
	}

	db.zabbixUser = mysqlUser
	if db.zabbixUser == "" {
		db.zabbixUser = "zabbix"
	}
	db.zabbixPassword = mysqlPassword
	if db.zabbixPassword == "" {
		db.zabbixPassword = "zabbix"
	}
	db.name = db.env.ValueOrDefaultNonEmpty("MYSQL_DATABASE", defaultDBName)

	return nil
}

// ApplyRuntimeEnvironment exports the resolved connection settings as
// ZBX_DB_* variables for the service.
func (db *Database) ApplyRuntimeEnvironment() {
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

	bootstrap.ApplyDatabaseCredentials(db.env, db.zabbixUser, db.zabbixPassword)
}

func (db *Database) Name() string     { return db.name }
func (db *Database) Schema() string   { return "" }
func (db *Database) User() string     { return db.zabbixUser }
func (db *Database) Password() string { return db.zabbixPassword }
