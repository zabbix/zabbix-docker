package postgresql

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/testutil"
)

type fakeSessionFactory struct {
	admin   *testutil.FakeDBSession
	schema  *testutil.FakeDBSession
	imports []*testutil.FakeDBSession
	configs []*pgx.ConnConfig
}

func (f *fakeSessionFactory) open(_ context.Context, config *pgx.ConnConfig) (dbSession, error) {
	f.configs = append(f.configs, config.Copy())
	if config.Database == "postgres" {
		return f.admin, nil
	}
	if config.User == "postgres" {
		return f.schema, nil
	}

	s := &testutil.FakeDBSession{Queries: map[string]string{}}
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

	db.ExportEnv()
	if host, found := env["ZBX_DB_HOST"]; !found || host != "" {
		t.Fatalf("ZBX_DB_HOST = %q, found = %t; want explicit empty value", host, found)
	}
}

func TestFrontendTLSConfigurationIsExplicit(t *testing.T) {
	env := bootstrap.Environment{"ZBX_DB_ENCRYPTION": "true"}

	service := NewForBackend(env)
	if err := service.Configure("zabbix"); err != nil {
		t.Fatal(err)
	}
	if service.tls.ConnectMode != "" {
		t.Fatalf("service used frontend TLS mode %q", service.tls.ConnectMode)
	}

	frontend := NewForFrontend(env)
	if err := frontend.Configure("zabbix"); err != nil {
		t.Fatal(err)
	}
	if frontend.tls.ConnectMode != "required" {
		t.Fatalf("frontend TLS mode = %q", frontend.tls.ConnectMode)
	}
}

func TestWaitForConnectionIsCanceled(t *testing.T) {
	db := NewForBackend(bootstrap.Environment{})
	if err := db.Configure("zabbix"); err != nil {
		t.Fatal(err)
	}

	db.open = func(ctx context.Context, _ *pgx.ConnConfig) (dbSession, error) {
		return nil, ctx.Err()
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := db.waitForConnectionContext(ctx, db.user, db.password)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("waitForConnectionContext() error = %v, want context.Canceled", err)
	}
}

func TestConfigureRejectsInvalidPort(t *testing.T) {
	err := NewForBackend(bootstrap.Environment{"DB_SERVER_PORT": "postgresql"}).Configure("zabbix")
	if err == nil || !strings.Contains(err.Error(), "DB_SERVER_PORT") {
		t.Fatalf("Configure() error = %v, want invalid DB_SERVER_PORT", err)
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
	testutil.WriteGzipFile(t, schemaFile, "CREATE TABLE dbversion (mandatory integer);")
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
		admin:  &testutil.FakeDBSession{Queries: map[string]string{}},
		schema: &testutil.FakeDBSession{Queries: map[string]string{}},
	}
	db.open = f.open
	if err := db.Prepare(schemaFile, timescaleFile); err != nil {
		t.Fatal(err)
	}

	adminStatements := strings.Join(f.admin.Statements, "\n")
	if !strings.Contains(adminStatements, `CREATE DATABASE "zabbix_test" OWNER "zabbix"`) {
		t.Fatalf("database was not created with the expected owner:\n%s", adminStatements)
	}
	schemaStatements := strings.Join(f.schema.Statements, "\n")
	for _, expected := range []string{
		`CREATE SCHEMA "monitoring" AUTHORIZATION "zabbix"`,
		"CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE",
	} {
		if !strings.Contains(schemaStatements, expected) {
			t.Fatalf("schema statements do not contain %q:\n%s", expected, schemaStatements)
		}
	}
	imports := testutil.ImportedStatements(f.imports)
	for _, expected := range []string{
		"CREATE TABLE dbversion (mandatory integer);",
		"SELECT 'timescale';",
		"SELECT 'extra';",
	} {
		if !strings.Contains(imports, expected) {
			t.Fatalf("import statements do not contain %q:\n%s", expected, imports)
		}
	}
	if !f.admin.Closed || !f.schema.Closed {
		t.Fatal("administrative PostgreSQL connections were not closed")
	}
	for _, s := range f.imports {
		if !s.Closed {
			t.Fatal("SQL import connection was not closed")
		}
	}
	logs := testutil.CaptureStdout(t, func() {
		if err := db.createSchema(schemaFile, timescaleFile); err != nil {
			t.Fatal(err)
		}
	})
	message := "Database schema successfully created!"
	if strings.Count(logs, message) != 1 || strings.Index(logs, message) < strings.Index(logs, "Processing additional") {
		t.Fatalf("expected one creation message after additional scripts:\n%s", logs)
	}

	if err := os.Mkdir(filepath.Join(scriptsDirectory, "20-invalid.sql"), 0o700); err != nil {
		t.Fatal(err)
	}
	logs = testutil.CaptureStdout(t, func() {
		if err := db.createSchema(schemaFile, timescaleFile); err == nil {
			t.Fatal("expected additional script failure")
		}
	})
	if strings.Contains(logs, message) {
		t.Fatal("creation success was logged after a script failure")
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

func TestExistingSchemaIsNotImported(t *testing.T) {
	db := NewForBackend(bootstrap.Environment{})
	if err := db.Configure("zabbix"); err != nil {
		t.Fatal(err)
	}
	existing := &testutil.FakeDBSession{Queries: map[string]string{
		"SELECT 1::text FROM pg_catalog.pg_class c JOIN pg_catalog.pg_namespace n ON n.oid=c.relnamespace WHERE c.relname='dbversion' AND n.nspname = $1": "1",
		"SELECT mandatory::text FROM \"public\".dbversion": "8000000",
	}}
	db.open = (&fakeSessionFactory{admin: existing, schema: existing}).open
	logs := testutil.CaptureStdout(t, func() {
		if err := db.createSchema("/does/not/exist.sql.gz", ""); err != nil {
			t.Fatal(err)
		}
	})
	if strings.Contains(logs, "Database schema successfully created!") {
		t.Fatal("creation success was logged for an existing schema")
	}
}
