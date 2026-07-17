package mysql

import (
	"context"
	"fmt"
	"strings"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

func (db *Database) executeSQLFile(path string) error {
	statements, err := readSQLStatements(path)
	if err != nil {
		return err
	}

	config, err := db.connectionConfig(db.name, db.rootUser, db.rootPassword)
	if err != nil {
		return err
	}
	sess, err := db.open(config)
	if err != nil {
		return fmt.Errorf("connect to database %s: %w", db.name, err)
	}
	defer sess.Close()

	for index, statement := range statements {
		if err := sess.Exec(context.Background(), statement); err != nil {
			return fmt.Errorf("execute SQL script %s statement %d: %w", path, index+1, err)
		}
	}

	return nil
}

func readSQLStatements(path string) ([]string, error) {
	data, err := bootstrap.ReadSQLFile(path)
	if err != nil {
		return nil, err
	}
	statements, err := splitSQLStatements(string(data))
	if err != nil {
		return nil, fmt.Errorf("parse SQL script %s: %w", path, err)
	}

	return statements, nil
}

func (db *Database) createDatabaseUser(sess databaseSession) error {
	if !db.createUser {
		return nil
	}

	bootstrap.LogInfo("** Creating '%s' user in MySQL database", db.zabbixUser)
	exists, err := db.query(sess, "SELECT 1 FROM mysql.user WHERE user = ? AND host = '%'", db.zabbixUser)
	if err != nil {
		return err
	}
	statement := "CREATE USER ?@'%' IDENTIFIED BY ?"
	if exists != "" {
		statement = "ALTER USER ?@'%' IDENTIFIED BY ?"
	}
	if err := db.execute(sess, statement, db.zabbixUser, db.zabbixPassword); err != nil {
		return err
	}

	return db.execute(sess, "GRANT "+zabbixDatabasePrivileges+" ON "+quoteIdentifier(db.name)+".* TO ?@'%'", db.zabbixUser)
}

func (db *Database) createDatabase(sess databaseSession) error {
	exists, err := db.query(sess, "SELECT SCHEMA_NAME FROM information_schema.SCHEMATA WHERE SCHEMA_NAME = ?", db.name)
	if err != nil {
		return err
	}
	if exists != "" {
		bootstrap.LogInfo("** Database '%s' already exists. Please be careful with database COLLATE!", db.name)
		return nil
	}

	bootstrap.LogInfo("** Database '%s' does not exist. Creating...", db.name)
	statement := "CREATE DATABASE " + quoteIdentifier(db.name) + " CHARACTER SET " + quoteIdentifier(db.characterSet) + " COLLATE " + quoteIdentifier(db.characterCollate)
	if err := db.execute(sess, statement); err != nil {
		return err
	}

	return db.execute(sess, "GRANT "+zabbixDatabasePrivileges+" ON "+quoteIdentifier(db.name)+".* TO ?@'%'", db.zabbixUser)
}

func (db *Database) createSchema(sess databaseSession, schemaFile string) error {
	exists, err := db.query(sess, "SELECT 1 FROM information_schema.tables WHERE table_schema = ? AND table_name = 'dbversion'", db.name)
	if err != nil {
		return err
	}

	version := ""
	if exists != "" {
		bootstrap.LogWarn("** Table '%s.dbversion' already exists.", db.name)
		version, err = db.query(sess, "SELECT mandatory FROM "+quoteIdentifier(db.name)+".dbversion")
		if err != nil {
			return err
		}
	}
	if version != "" {
		return nil
	}

	bootstrap.LogInfo("** Creating '%s' schema in MySQL", db.name)
	if err := db.executeSQLFile(schemaFile); err != nil {
		return err
	}
	bootstrap.LogInfo("** Database schema successfully created!")

	return bootstrap.RunAdditionalSQLScripts(db.env, db.executeSQLFile)
}

// Prepare provisions the database over the administrative connection: it
// creates the Zabbix user and database when missing and imports the schema
// together with the additional dbscripts.
func (db *Database) Prepare(schemaFile string) (err error) {
	admin, err := db.waitForConnection(db.rootUser, db.rootPassword)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := admin.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("close database connection: %w", closeErr)
		}
	}()

	if err := db.createDatabaseUser(admin); err != nil {
		return err
	}
	if err := db.createDatabase(admin); err != nil {
		return err
	}

	return db.createSchema(admin, schemaFile)
}

func quoteIdentifier(value string) string {
	return "`" + strings.ReplaceAll(value, "`", "``") + "`"
}
