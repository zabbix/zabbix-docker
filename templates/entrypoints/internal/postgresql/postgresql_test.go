package postgresql

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

type fakeDBSession struct {
	queries    map[string]string
	statements []string
	closed     bool
}

func (s *fakeDBSession) QueryString(_ context.Context, query string, args ...any) (string, error) {
	s.statements = append(s.statements, formatStatement(query, args...))
	return s.queries[query], nil
}

func (s *fakeDBSession) Exec(_ context.Context, query string, args ...any) error {
	s.statements = append(s.statements, formatStatement(query, args...))
	return nil
}

func (s *fakeDBSession) Close(context.Context) error {
	s.closed = true
	return nil
}

type fakeSessionFactory struct {
	admin   *fakeDBSession
	schema  *fakeDBSession
	imports []*fakeDBSession
	configs []*pgx.ConnConfig
}

func (f *fakeSessionFactory) connect(_ context.Context, config *pgx.ConnConfig) (dbSession, error) {
	f.configs = append(f.configs, config.Copy())
	if config.Database == "postgres" {
		return f.admin, nil
	}
	if config.User == "postgres" {
		return f.schema, nil
	}

	s := &fakeDBSession{queries: map[string]string{}}
	f.imports = append(f.imports, s)
	return s, nil
}

func TestConfigureSocketAndDefaults(t *testing.T) {
	env := bootstrap.Environment{"DB_SERVER_HOST": ""}
	db := NewForBackend(env)
	if err := db.Configure("zabbix"); err != nil {
		t.Fatal(err)
	}
	if db.host != "" || db.port != "5432" || db.schema != "public" {
		t.Fatalf("unexpected connection settings: %#v", db)
	}
	if db.user != "zabbix" || db.password != "zabbix" || db.adminUser != "postgres" {
		t.Fatalf("unexpected credentials: %#v", db)
	}
}

func TestFrontendTLSConfigurationIsExplicit(t *testing.T) {
	env := bootstrap.Environment{"ZBX_DB_ENCRYPTION": "true"}

	service := NewForBackend(env)
	if service.tls.ConnectMode != "" {
		t.Fatalf("service used frontend TLS mode %q", service.tls.ConnectMode)
	}

	frontend := NewForFrontend(env)
	if frontend.tls.ConnectMode != "required" {
		t.Fatalf("frontend TLS mode = %q", frontend.tls.ConnectMode)
	}
}

func TestWaitForConnectionIsCanceled(t *testing.T) {
	db := NewForBackend(bootstrap.Environment{})
	if err := db.Configure("zabbix"); err != nil {
		t.Fatal(err)
	}

	db.connect = func(ctx context.Context, _ *pgx.ConnConfig) (dbSession, error) {
		return nil, ctx.Err()
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := db.waitForConnectionContext(ctx, db.user, db.password)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("waitForConnectionContext() error = %v, want context.Canceled", err)
	}
}

func TestPrepareDatabase(t *testing.T) {
	root := t.TempDir()
	scriptsDirectory := filepath.Join(root, "dbscripts")
	if err := os.Mkdir(scriptsDirectory, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(scriptsDirectory, "10-extra.sql"), []byte("SELECT 'extra';"), 0o600); err != nil {
		t.Fatal(err)
	}
	schemaFile := filepath.Join(root, "create.sql.gz")
	writeGzipFile(t, schemaFile, "CREATE TABLE dbversion (mandatory integer);")
	timescaleFile := filepath.Join(root, "timescaledb.sql")
	if err := os.WriteFile(timescaleFile, []byte("SELECT 'timescale';"), 0o600); err != nil {
		t.Fatal(err)
	}

	env := bootstrap.Environment{
		"DB_SERVER_HOST":       "database",
		"POSTGRES_USER":        "postgres",
		"POSTGRES_PASSWORD":    "root-password",
		"POSTGRES_DB":          "zabbix_test",
		"DB_SERVER_SCHEMA":     "monitoring",
		"ENABLE_TIMESCALEDB":   "true",
		"ZABBIX_USER_HOME_DIR": root,
	}
	db := NewForBackend(env)
	if err := db.Configure("zabbix"); err != nil {
		t.Fatal(err)
	}
	db.user = "zabbix"
	db.password = "zabbix-password"

	f := &fakeSessionFactory{
		admin:  &fakeDBSession{queries: map[string]string{}},
		schema: &fakeDBSession{queries: map[string]string{}},
	}
	db.connect = f.connect
	if err := db.Prepare(schemaFile, timescaleFile); err != nil {
		t.Fatal(err)
	}

	adminStatements := strings.Join(f.admin.statements, "\n")
	if !strings.Contains(adminStatements, `CREATE DATABASE "zabbix_test" OWNER "zabbix"`) {
		t.Fatalf("database was not created with the expected owner:\n%s", adminStatements)
	}
	schemaStatements := strings.Join(f.schema.statements, "\n")
	for _, expected := range []string{
		`CREATE SCHEMA "monitoring" AUTHORIZATION "zabbix"`,
		"CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE",
	} {
		if !strings.Contains(schemaStatements, expected) {
			t.Fatalf("schema statements do not contain %q:\n%s", expected, schemaStatements)
		}
	}
	imports := importedStatements(f.imports)
	for _, expected := range []string{
		"CREATE TABLE dbversion (mandatory integer);",
		"SELECT 'timescale';",
		"SELECT 'extra';",
	} {
		if !strings.Contains(imports, expected) {
			t.Fatalf("import statements do not contain %q:\n%s", expected, imports)
		}
	}
	if !f.admin.closed || !f.schema.closed {
		t.Fatal("administrative PostgreSQL connections were not closed")
	}
	for _, s := range f.imports {
		if !s.closed {
			t.Fatal("SQL import connection was not closed")
		}
	}
}

func TestQuoteIdentifier(t *testing.T) {
	if quoted := quoteIdentifier(`schema"name`); quoted != `"schema""name"` {
		t.Fatalf("unexpected quoted identifier: %s", quoted)
	}
}

func TestExportEnvOmitsVaultCredentials(t *testing.T) {
	env := bootstrap.Environment{
		"ZBX_DB_USER": "old-user", "ZBX_DB_PASSWORD": "old-password",
	}
	db := &DB{
		env: env, name: "zabbix", user: "vault-user", password: "vault-password", fromVault: true,
	}
	db.ExportEnv()
	if _, found := env["ZBX_DB_USER"]; found {
		t.Fatalf("Vault credentials were exported: %#v", env)
	}
	if _, found := env["ZBX_DB_PASSWORD"]; found {
		t.Fatalf("Vault credentials were exported: %#v", env)
	}
}

func TestExportEnv(t *testing.T) {
	env := bootstrap.Environment{
		"DB_SERVER_HOST":   "postgres-server",
		"DB_SERVER_PORT":   "5432",
		"DB_SERVER_SOCKET": "/run/postgresql",
	}
	db := &DB{
		env: env, name: "zabbix", schema: "monitoring", user: "zabbix", password: "secret",
	}

	db.ExportEnv()

	expected := map[string]string{
		"ZBX_DB_HOST":     "postgres-server",
		"ZBX_DB_PORT":     "5432",
		"ZBX_DB_NAME":     "zabbix",
		"ZBX_DB_SCHEMA":   "monitoring",
		"ZBX_DB_USER":     "zabbix",
		"ZBX_DB_PASSWORD": "secret",
	}
	for name, value := range expected {
		if env[name] != value {
			t.Fatalf("%s = %q, want %q", name, env[name], value)
		}
	}
	if _, found := env["ZBX_DB_SOCKET"]; found {
		t.Fatalf("PostgreSQL runtime environment contains a MySQL socket: %#v", env)
	}
}

func writeGzipFile(t *testing.T, path, content string) {
	t.Helper()

	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	compressed := gzip.NewWriter(file)
	if _, err := compressed.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := compressed.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}

func formatStatement(query string, args ...any) string {
	if len(args) == 0 {
		return query
	}
	return fmt.Sprintf("%s %v", query, args)
}

func importedStatements(sessions []*fakeDBSession) string {
	var statements []string
	for _, s := range sessions {
		statements = append(statements, s.statements...)
	}
	return strings.Join(statements, "\n")
}
