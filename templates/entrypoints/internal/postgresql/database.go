// Package postgresql provisions the PostgreSQL database for Zabbix server
// and the web frontend images.
package postgresql

import (
	"cmp"
	"strings"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/vault"
)

// DB carries the resolved PostgreSQL connection target and the
// working and administrative credentials.
type DB struct {
	env                bootstrap.Environment
	tls                bootstrap.DBTLSConfig
	open               sessionOpener
	host               string
	port               string
	adminUser          string
	adminPass          string
	user               string
	password           string
	name               string
	schema             string
	implicitSearchPath bool
	fromVault          bool
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
	return &DB{env: env, tls: tls, open: openDBSession}
}

// Configure resolves the connection target, schema and credentials from the
// environment or Vault.
func (db *DB) Configure(defaultDBName string) error {
	host := db.env.ValueOrDefault("DB_SERVER_HOST", "postgres-server")
	port := db.env.ValueOrDefaultNonEmpty("DB_SERVER_PORT", "5432")
	db.env["DB_SERVER_HOST"] = host
	db.env["DB_SERVER_PORT"] = port

	schema := db.env.ValueOrDefault("DB_SERVER_SCHEMA", "public")
	db.env["DB_SERVER_SCHEMA"] = schema

	db.host = host
	db.port = port
	db.schema = schema

	creds, err := vault.ResolveDBCredentials(db.env)
	if err != nil {
		return err
	}

	for _, name := range []string{"POSTGRES_USER", "POSTGRES_PASSWORD"} {
		if err := bootstrap.ResolveSecretEnv(db.env, name); err != nil {
			return err
		}
	}

	postgresUser := db.env["POSTGRES_USER"]
	postgresPassword := db.env["POSTGRES_PASSWORD"]
	db.adminUser = cmp.Or(postgresUser, "postgres")
	db.adminPass = postgresPassword
	db.user = cmp.Or(postgresUser, "zabbix")
	db.password = cmp.Or(postgresPassword, "zabbix")
	if creds != nil {
		db.user = creds.Username
		db.password = creds.Password
		if postgresUser == "" && postgresPassword == "" {
			db.adminUser = creds.Username
			db.adminPass = creds.Password
		}
	}
	db.name = db.env.ValueOrDefaultNonEmpty("POSTGRES_DB", defaultDBName)
	db.implicitSearchPath = strings.EqualFold(
		db.env.ValueOrDefaultNonEmpty("POSTGRES_USE_IMPLICIT_SEARCH_PATH", "false"), "true",
	)
	db.fromVault = creds != nil

	return nil
}

// ExportEnv exports the resolved connection settings as
// ZBX_DB_* variables for the service.
func (db *DB) ExportEnv() {
	if db.env["DB_SERVER_HOST"] != "" {
		db.env["ZBX_DB_HOST"] = db.env["DB_SERVER_HOST"]
	}
	if db.env["DB_SERVER_PORT"] != "" {
		db.env["ZBX_DB_PORT"] = db.env["DB_SERVER_PORT"]
	}
	db.env["ZBX_DB_NAME"] = db.name
	if db.schema != "" {
		db.env["ZBX_DB_SCHEMA"] = db.schema
	}

	if db.fromVault {
		delete(db.env, "ZBX_DB_USER")
		delete(db.env, "ZBX_DB_PASSWORD")
	} else {
		db.env["ZBX_DB_USER"] = db.user
		db.env["ZBX_DB_PASSWORD"] = db.password
	}
}

func (db *DB) Name() string     { return db.name }
func (db *DB) Schema() string   { return db.schema }
func (db *DB) User() string     { return db.user }
func (db *DB) Password() string { return db.password }
