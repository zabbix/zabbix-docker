package sqlite

import (
	"compress/gzip"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func TestPrepareInitializesEmptyDatabase(t *testing.T) {
	directory := t.TempDir()
	databasePath := filepath.Join(directory, "proxy.db")
	if err := os.WriteFile(databasePath, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	schemaPath := filepath.Join(directory, "schema.sql.gz")
	file, err := os.Create(schemaPath)
	if err != nil {
		t.Fatal(err)
	}
	compressed := gzip.NewWriter(file)
	if _, err := compressed.Write([]byte("CREATE TABLE dbversion (mandatory INTEGER); INSERT INTO dbversion VALUES (8000000);")); err != nil {
		t.Fatal(err)
	}
	if err := compressed.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	if err := Prepare(databasePath, schemaPath); err != nil {
		t.Fatal(err)
	}
	database, err := sql.Open("sqlite3", "file:"+databasePath)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	var version int
	if err := database.QueryRow("SELECT mandatory FROM dbversion").Scan(&version); err != nil {
		t.Fatal(err)
	}
	if version != 8000000 {
		t.Fatalf("unexpected version: %d", version)
	}
	if err := Prepare(databasePath, schemaPath); err != nil {
		t.Fatalf("second preparation must be idempotent: %v", err)
	}
}

func TestPrepareRejectsInvalidDatabase(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "proxy.db")
	if err := os.WriteFile(databasePath, []byte("not sqlite"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := Prepare(databasePath, "unused.sql.gz"); err == nil {
		t.Fatal("expected invalid database error")
	}
}

func TestPrepareEscapesDatabasePath(t *testing.T) {
	directory := t.TempDir()
	databasePath := filepath.Join(directory, "proxy?#.db")
	schemaPath := filepath.Join(directory, "schema.sql.gz")
	writeSchema(t, schemaPath)

	if err := Prepare(databasePath, schemaPath); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(databasePath); err != nil {
		t.Fatalf("database was not created at the requested path: %v", err)
	}
}

func writeSchema(t *testing.T, path string) {
	t.Helper()

	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	compressed := gzip.NewWriter(file)
	if _, err := compressed.Write([]byte("CREATE TABLE dbversion (mandatory INTEGER); INSERT INTO dbversion VALUES (8000000);")); err != nil {
		t.Fatal(err)
	}
	if err := compressed.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}
