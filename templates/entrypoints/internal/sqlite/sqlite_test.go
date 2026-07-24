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

func TestPrepareRejectsIncompleteExistingSchema(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "proxy.db")
	database, err := sql.Open("sqlite3", "file:"+databasePath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec("CREATE TABLE dbversion (mandatory INTEGER)"); err != nil {
		t.Fatal(err)
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}

	if err := Prepare(databasePath, "unused.sql.gz"); err == nil {
		t.Fatal("expected incomplete schema error")
	}
}

func TestPrepareRollsBackFailedSchema(t *testing.T) {
	directory := t.TempDir()
	databasePath := filepath.Join(directory, "proxy.db")
	schemaPath := filepath.Join(directory, "schema.sql.gz")
	writeSQLSchema(t, schemaPath,
		"CREATE TABLE dbversion (mandatory INTEGER); "+
			"INSERT INTO dbversion VALUES (8000000); "+
			"INVALID SQL;",
	)

	if err := Prepare(databasePath, schemaPath); err == nil {
		t.Fatal("expected schema creation error")
	}

	database, err := sql.Open("sqlite3", "file:"+databasePath)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	var tables int
	if err := database.QueryRow(
		"SELECT count(*) FROM sqlite_master WHERE type='table' AND name='dbversion'",
	).Scan(&tables); err != nil {
		t.Fatal(err)
	}
	if tables != 0 {
		t.Fatal("dbversion table was not rolled back")
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

	writeSQLSchema(t, path, "CREATE TABLE dbversion (mandatory INTEGER); INSERT INTO dbversion VALUES (8000000);")
}

func writeSQLSchema(t *testing.T, path, schema string) {
	t.Helper()

	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	compressed := gzip.NewWriter(file)
	if _, err := compressed.Write([]byte(schema)); err != nil {
		t.Fatal(err)
	}
	if err := compressed.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}
