// Package postgresql provisions the PostgreSQL database for Zabbix server
// and the web frontend images.
package postgresql

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

// Database carries the resolved PostgreSQL connection target and the
// working and administrative credentials.
type Database struct {
	env      bootstrap.Environment
	connect  connector
	host     string
	port     uint16
	rootUser string
	rootPass string
	user     string
	password string
	name     string
	schema   string
	implicit bool
}

// New creates an unconfigured Database; call Configure before use.
func New(env bootstrap.Environment) *Database {
	return &Database{env: env, connect: openDatabaseSession}
}

// Configure resolves the connection target, schema and credentials from
// the environment: POSTGRES_* variables with the *_FILE secret convention,
// or credentials coming from Vault.
func (db *Database) Configure(defaultDBName string, creds *bootstrap.DBCredentials) error {
	host, found := db.env["DB_SERVER_HOST"]
	if !found {
		host = "postgres-server"
		db.env["DB_SERVER_HOST"] = host
	}

	portValue := db.env.ValueOrDefaultNonEmpty("DB_SERVER_PORT", "5432")
	port, err := strconv.ParseUint(portValue, 10, 16)
	if err != nil {
		return fmt.Errorf("invalid DB_SERVER_PORT %q: %w", portValue, err)
	}
	db.env["DB_SERVER_PORT"] = portValue

	schema, found := db.env["DB_SERVER_SCHEMA"]
	if !found {
		schema = "public"
		db.env["DB_SERVER_SCHEMA"] = schema
	}

	for _, name := range []string{"POSTGRES_USER", "POSTGRES_PASSWORD"} {
		if err := bootstrap.FileEnv(db.env, name, ""); err != nil {
			return err
		}
	}

	db.host = host
	db.port = uint16(port)
	db.rootUser = db.env.ValueOrDefaultNonEmpty("POSTGRES_USER", "postgres")
	db.rootPass = db.env["POSTGRES_PASSWORD"]
	db.user = db.env.ValueOrDefaultNonEmpty("POSTGRES_USER", "zabbix")
	db.password = db.env.ValueOrDefaultNonEmpty("POSTGRES_PASSWORD", "zabbix")
	if creds != nil {
		db.user = creds.Username
		db.password = creds.Password
		if db.env["POSTGRES_USER"] == "" && db.env["POSTGRES_PASSWORD"] == "" {
			db.rootUser = creds.Username
			db.rootPass = creds.Password
		}
	}
	db.name = db.env.ValueOrDefaultNonEmpty("POSTGRES_DB", defaultDBName)
	db.schema = schema
	db.implicit = strings.EqualFold(db.env.ValueOrDefaultNonEmpty("POSTGRES_USE_IMPLICIT_SEARCH_PATH", "false"), "true")

	return nil
}

// ApplyRuntimeEnvironment exports the resolved connection settings as
// ZBX_DB_* variables for the service.
func (db *Database) ApplyRuntimeEnvironment() {
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

	bootstrap.ApplyDatabaseCredentials(db.env, db.user, db.password)
}

func (db *Database) Name() string     { return db.name }
func (db *Database) Schema() string   { return db.schema }
func (db *Database) User() string     { return db.user }
func (db *Database) Password() string { return db.password }
