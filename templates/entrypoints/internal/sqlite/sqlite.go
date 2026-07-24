// Package sqlite prepares the SQLite database file for Zabbix proxy.
package sqlite

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	_ "github.com/ncruces/go-sqlite3/driver"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

// Prepare makes sure a valid Zabbix database exists at dbPath,
// creating the file and importing the (optionally gzipped) schema on first
// start. An existing database is left untouched.
func Prepare(dbPath, schemaPath string) error {
	if dbPath == "" {
		return fmt.Errorf("ZBX_DB_NAME must be set")
	}

	directory := filepath.Dir(dbPath)
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		bootstrap.LogInfo("** SQLite database directory '%s' does not exist. Creating...", directory)
		if err := os.MkdirAll(directory, 0o775); err != nil {
			return err
		}
	}
	_, statErr := os.Stat(dbPath)
	existed := statErr == nil
	db, err := sql.Open("sqlite3", dbDSN(dbPath))
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		if existed {
			return fmt.Errorf("file %q exists but is not a valid SQLite database: %w", dbPath, err)
		}
		return err
	}
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return fmt.Errorf("enable SQLite WAL: %w", err)
	}

	var result int
	if err := db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='dbversion'").Scan(&result); err != nil {
		return err
	}

	if result == 0 {
		bootstrap.LogInfo("** SQLite database '%s' does not contain Zabbix schema. Creating...", dbPath)
		data, err := bootstrap.ReadSQLFile(schemaPath)
		if err != nil {
			return err
		}
		if err := createSchema(db, data); err != nil {
			return err
		}
	} else {
		var version int
		if err := db.QueryRow("SELECT mandatory FROM dbversion LIMIT 1").Scan(&version); err != nil {
			return fmt.Errorf("validate SQLite schema in %q: %w", dbPath, err)
		}
		bootstrap.LogInfo("** SQLite database '%s' already exists", dbPath)
	}

	return nil
}

func createSchema(db *sql.DB, data []byte) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin SQLite schema transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.Exec(string(data)); err != nil {
		return fmt.Errorf("create SQLite schema: %w", err)
	}

	var version int
	if err := tx.QueryRow("SELECT mandatory FROM dbversion LIMIT 1").Scan(&version); err != nil {
		return fmt.Errorf("validate created SQLite schema: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit SQLite schema: %w", err)
	}

	return nil
}

func dbDSN(dbPath string) string {
	dsn := &url.URL{
		Scheme:   "file",
		Path:     filepath.ToSlash(dbPath),
		RawQuery: url.Values{"_txlock": {"immediate"}}.Encode(),
	}

	return dsn.String()
}
