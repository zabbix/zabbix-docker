// Package mysql provisions the MySQL database for Zabbix server, proxy and
// the web frontend images.
package mysql

import (
	"fmt"
	"net"
	"strings"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
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
	rootUser   string
	rootPass   string
	user       string
	password   string
	name       string
	createUser bool
}

// New creates an unconfigured DB; call Configure before use.
func New(env bootstrap.Environment) *DB {
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

// Configure resolves the connection target and the credentials from the
// environment, following the same rules as the shell entrypoints did:
// MYSQL_* variables with the *_FILE secret convention, optional root
// access and credentials coming from Vault.
func (db *DB) Configure(defaultDBName string, creds *bootstrap.DBCredentials) error {
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
		db.rootPass = mysqlRootPassword
	} else {
		db.rootUser = db.env["DB_SERVER_ROOT_USER"]
		db.rootPass = db.env["DB_SERVER_ROOT_PASS"]
	}

	db.createUser = creds == nil && mysqlUser != "" && useRootUser
	if db.rootUser == "" {
		db.rootUser = mysqlUser
	}
	if !allowEmptyPassword && db.rootPass == "" {
		db.rootPass = mysqlPassword
	}

	db.user = mysqlUser
	if db.user == "" {
		db.user = "zabbix"
	}
	db.password = mysqlPassword
	if db.password == "" {
		db.password = "zabbix"
	}
	db.name = db.env.ValueOrDefaultNonEmpty("MYSQL_DATABASE", defaultDBName)

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

	bootstrap.ApplyDBCredentials(db.env, db.user, db.password)
}

func (db *DB) Name() string     { return db.name }
func (db *DB) Schema() string   { return "" }
func (db *DB) User() string     { return db.user }
func (db *DB) Password() string { return db.password }
