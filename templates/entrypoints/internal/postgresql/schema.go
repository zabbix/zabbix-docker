package postgresql

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

func (db *Database) createDatabase(sess databaseSession) error {
	exists, err := sess.QueryString(context.Background(), "SELECT 1::text FROM pg_database WHERE datname=$1", db.name)
	if err != nil {
		return err
	}
	if exists != "" {
		bootstrap.LogInfo("** Database '%s' already exists. Please be careful with database owner!", db.name)
		return nil
	}

	bootstrap.LogInfo("** Database '%s' does not exist. Creating...", db.name)
	statement := "CREATE DATABASE " + quoteIdentifier(db.name) + " OWNER " + quoteIdentifier(db.user) + " LC_CTYPE 'en_US.utf8' LC_COLLATE 'en_US.utf8'"
	if err := sess.Exec(context.Background(), statement); err != nil {
		return fmt.Errorf("create PostgreSQL database: %w", err)
	}

	return nil
}

func (db *Database) createSchemaNamespace(sess databaseSession) error {
	if db.schema == "" {
		return nil
	}

	exists, err := sess.QueryString(context.Background(), "SELECT 1::text FROM pg_namespace WHERE nspname=$1", db.schema)
	if err != nil {
		return err
	}
	if exists == "" {
		return sess.Exec(context.Background(), "CREATE SCHEMA "+quoteIdentifier(db.schema)+" AUTHORIZATION "+quoteIdentifier(db.user))
	}

	return nil
}

func (db *Database) executeSQLFile(path string) error {
	data, err := readSQLFile(path)
	if err != nil {
		return err
	}

	sess, err := db.targetConnection(db.user, db.password)
	if err != nil {
		return fmt.Errorf("connect to PostgreSQL database %s: %w", db.name, err)
	}
	defer sess.Close(context.Background())
	if err := sess.Exec(context.Background(), string(data)); err != nil {
		return fmt.Errorf("execute SQL script %s: %w", path, err)
	}

	return nil
}

func (db *Database) createSchema(schemaFile, timescaleDBSchemaFile string) error {
	sess, err := db.targetConnection(db.rootUser, db.rootPass)
	if err != nil {
		return err
	}
	defer sess.Close(context.Background())

	if err := db.createSchemaNamespace(sess); err != nil {
		return err
	}

	filter := ""
	table := "dbversion"
	if db.schema != "" {
		filter = " AND n.nspname = $1"
		table = quoteIdentifier(db.schema) + ".dbversion"
	}
	query := "SELECT 1::text FROM pg_catalog.pg_class c JOIN pg_catalog.pg_namespace n ON n.oid=c.relnamespace WHERE c.relname='dbversion'" + filter
	args := []any{}
	if db.schema != "" {
		args = append(args, db.schema)
	}
	exists, err := sess.QueryString(context.Background(), query, args...)
	if err != nil {
		return err
	}

	version := ""
	if exists != "" {
		bootstrap.LogInfo("** Table '%s.dbversion' already exists.", db.name)
		version, err = sess.QueryString(context.Background(), "SELECT mandatory::text FROM "+table)
		if err != nil {
			return err
		}
	}
	if version != "" {
		return nil
	}

	bootstrap.LogInfo("** Creating '%s' schema in PostgreSQL", db.name)
	if err := db.executeSQLFile(schemaFile); err != nil {
		return err
	}
	if strings.EqualFold(db.env["ENABLE_TIMESCALEDB"], "true") {
		if err := sess.Exec(context.Background(), "CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE"); err != nil {
			return err
		}
		if err := db.executeSQLFile(timescaleDBSchemaFile); err != nil {
			return err
		}
	}

	homeDir, err := bootstrap.RequiredHomeDirectory(db.env)
	if err != nil {
		return err
	}
	scripts, err := filepath.Glob(filepath.Join(homeDir, "dbscripts", "*.sql"))
	if err != nil {
		return fmt.Errorf("find additional SQL scripts: %w", err)
	}
	for _, script := range scripts {
		bootstrap.LogInfo("** Processing additional '%s' SQL script", script)
		if err := db.executeSQLFile(script); err != nil {
			return err
		}
	}

	return nil
}

// Prepare provisions the database over the administrative connection: it
// creates the Zabbix database when missing and imports the schema,
// enabling TimescaleDB support when available.
func (db *Database) Prepare(schemaFile, timescaleDBSchemaFile string) error {
	sess, err := db.waitForConnection(db.rootUser, db.rootPass)
	if err != nil {
		return err
	}
	if err := db.createDatabase(sess); err != nil {
		_ = sess.Close(context.Background())
		return err
	}
	if err := sess.Close(context.Background()); err != nil {
		return fmt.Errorf("close PostgreSQL connection: %w", err)
	}

	return db.createSchema(schemaFile, timescaleDBSchemaFile)
}

func quoteIdentifier(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

func readSQLFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open SQL script %s: %w", path, err)
	}
	defer file.Close()

	var reader io.Reader = file
	if strings.HasSuffix(path, ".gz") {
		compressed, err := gzip.NewReader(file)
		if err != nil {
			return nil, err
		}
		defer compressed.Close()
		reader = compressed
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read SQL script %s: %w", path, err)
	}

	return data, nil
}
