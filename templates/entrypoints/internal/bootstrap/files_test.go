package bootstrap

import (
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestReadSQLFile(t *testing.T) {
	dir := t.TempDir()

	plain := filepath.Join(dir, "schema.sql")
	if err := os.WriteFile(plain, []byte("SELECT 1;"), 0o644); err != nil {
		t.Fatal(err)
	}

	var buffer bytes.Buffer
	writer := gzip.NewWriter(&buffer)
	if _, err := writer.Write([]byte("SELECT 2;")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	compressed := filepath.Join(dir, "schema.sql.gz")
	if err := os.WriteFile(compressed, buffer.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}

	data, err := ReadSQLFile(plain)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "SELECT 1;" {
		t.Fatalf("plain file content = %q", data)
	}

	data, err = ReadSQLFile(compressed)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "SELECT 2;" {
		t.Fatalf("compressed file content = %q", data)
	}

	if _, err := ReadSQLFile(filepath.Join(dir, "missing.sql")); err == nil {
		t.Fatal("missing file was accepted")
	}
	if _, err := ReadSQLFile(plain); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadSQLFile(filepath.Join(dir, "absent.sql.gz")); err == nil {
		t.Fatal("missing compressed file was accepted")
	}
	notGzip := filepath.Join(dir, "broken.sql.gz")
	if err := os.WriteFile(notGzip, []byte("not a gzip"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadSQLFile(notGzip); err == nil {
		t.Fatal("invalid gzip content was accepted")
	}
}

func TestRunAdditionalSQLScripts(t *testing.T) {
	home := t.TempDir()
	env := Environment{"ZABBIX_USER_HOME_DIR": home}

	if err := RunAdditionalSQLScripts(env, func(string) error {
		t.Fatal("script executed although dbscripts directory does not exist")
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	scriptsDir := filepath.Join(home, "dbscripts")
	if err := os.Mkdir(scriptsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"b.sql", "a.sql", "ignored.txt"} {
		if err := os.WriteFile(filepath.Join(scriptsDir, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	var executed []string
	if err := RunAdditionalSQLScripts(env, func(path string) error {
		executed = append(executed, filepath.Base(path))
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if len(executed) != 2 || executed[0] != "a.sql" || executed[1] != "b.sql" {
		t.Fatalf("executed scripts = %v", executed)
	}
}
