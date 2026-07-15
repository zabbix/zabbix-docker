// Package sqlite prepares the SQLite database file for Zabbix proxy.
package sqlite

import (
	"compress/gzip"
	"database/sql"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

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

	var result int
	if err := db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='dbversion'").Scan(&result); err != nil {
		return err
	}

	if result == 0 {
		bootstrap.LogInfo("** SQLite database '%s' does not contain Zabbix schema. Creating...", dbPath)
		data, err := readSchema(schemaPath)
		if err != nil {
			return err
		}
		if _, err := db.Exec(string(data)); err != nil {
			return fmt.Errorf("create SQLite schema: %w", err)
		}
		if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
			return fmt.Errorf("enable SQLite WAL: %w", err)
		}
	} else {
		bootstrap.LogInfo("** SQLite database '%s' already exists.", dbPath)
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

func readSchema(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
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

	return io.ReadAll(reader)
}
