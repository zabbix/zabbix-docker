package mysql

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

type fakeDatabaseSession struct {
	queries    map[string]string
	statements []string
	closed     bool
}

func (s *fakeDatabaseSession) Ping(ctx context.Context) error {
	return ctx.Err()
}

func (s *fakeDatabaseSession) QueryString(_ context.Context, query string, args ...any) (string, error) {
	s.statements = append(s.statements, formatStatement(query, args...))
	return s.queries[query], nil
}

func (s *fakeDatabaseSession) Exec(_ context.Context, query string, args ...any) error {
	s.statements = append(s.statements, formatStatement(query, args...))
	return nil
}

func (s *fakeDatabaseSession) Close() error {
	s.closed = true
	return nil
}

type fakeSessionFactory struct {
	admin   *fakeDatabaseSession
	imports []*fakeDatabaseSession
	configs []*mysql.Config
}

func (f *fakeSessionFactory) open(config *mysql.Config) (databaseSession, error) {
	f.configs = append(f.configs, config.Clone())
	if config.DBName == "" {
		return f.admin, nil
	}
	s := &fakeDatabaseSession{queries: map[string]string{}}
	f.imports = append(f.imports, s)
	return s, nil
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

	env := bootstrap.Environment{
		"DB_SERVER_HOST": "database", "MYSQL_USER": "zabbix",
		"MYSQL_PASSWORD": "zabbix-password", "MYSQL_ROOT_PASSWORD": "root-password",
		"MYSQL_DATABASE": "zabbix_test", "ZABBIX_USER_HOME_DIR": root,
	}
	database := New(env)
	f := &fakeSessionFactory{admin: &fakeDatabaseSession{queries: map[string]string{}}}
	database.open = f.open
	if err := database.Configure("zabbix", nil); err != nil {
		t.Fatal(err)
	}
	if err := database.Prepare(schemaFile); err != nil {
		t.Fatal(err)
	}

	adminStatements := strings.Join(f.admin.statements, "\n")
	for _, expected := range []string{
		"CREATE USER ?@'%' IDENTIFIED BY ? [zabbix zabbix-password]",
		"CREATE DATABASE `zabbix_test` CHARACTER SET `utf8mb4` COLLATE `utf8mb4_bin`",
		"GRANT " + zabbixDatabasePrivileges + " ON `zabbix_test`.* TO ?@'%' [zabbix]",
	} {
		if !strings.Contains(adminStatements, expected) {
			t.Fatalf("admin statements do not contain %q:\n%s", expected, adminStatements)
		}
	}
	if strings.Contains(adminStatements, "GRANT ALL PRIVILEGES") {
		t.Fatalf("admin statements contain an unrestricted grant:\n%s", adminStatements)
	}
	imports := importedStatements(f.imports)
	for _, expected := range []string{"CREATE TABLE dbversion (mandatory integer)", "SELECT 'extra'"} {
		if !strings.Contains(imports, expected) {
			t.Fatalf("import statements do not contain %q:\n%s", expected, imports)
		}
	}
	if len(f.configs) == 0 || f.configs[0].Passwd != "root-password" {
		t.Fatalf("root password was not configured: %#v", f.configs)
	}
	if !f.admin.closed {
		t.Fatal("admin connection was not closed")
	}
}

func TestRequiredTLSConfiguration(t *testing.T) {
	env := bootstrap.Environment{"ZBX_DBTLSCONNECT": "required"}
	database := New(env)
	config, err := database.tlsConfig()
	if err != nil {
		t.Fatal(err)
	}
	if config == nil || !config.InsecureSkipVerify {
		t.Fatalf("required TLS config = %#v", config)
	}
}

func TestWaitForConnectionIsCanceled(t *testing.T) {
	database := New(bootstrap.Environment{"MYSQL_ALLOW_EMPTY_PASSWORD": "true"})
	if err := database.Configure("zabbix", nil); err != nil {
		t.Fatal(err)
	}

	database.open = func(*mysql.Config) (databaseSession, error) {
		return &fakeDatabaseSession{}, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := database.waitForConnectionContext(ctx, database.zabbixUser, database.zabbixPassword)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("waitForConnectionContext() error = %v, want context.Canceled", err)
	}
}

func TestVaultCredentialsAreUsedWithoutAdministrativeCredentials(t *testing.T) {
	db := New(bootstrap.Environment{})
	creds := &bootstrap.DBCredentials{Username: "vault-user", Password: "vault-password"}
	if err := db.Configure("zabbix", creds); err != nil {
		t.Fatal(err)
	}
	if db.zabbixUser != "vault-user" || db.zabbixPassword != "vault-password" {
		t.Fatal("Vault credentials were not configured for Zabbix")
	}
	if db.rootUser != "vault-user" || db.rootPassword != "vault-password" {
		t.Fatal("Vault credentials were not used as administrative fallback")
	}
}

func TestApplyRuntimeEnvironment(t *testing.T) {
	env := bootstrap.Environment{
		"DB_SERVER_SOCKET": "/run/mysqld/mysqld.sock",
		"DB_SERVER_HOST":   "mysql-server",
		"DB_SERVER_PORT":   "3306",
	}
	database := &Database{
		env: env, name: "zabbix", zabbixUser: "zabbix", zabbixPassword: "secret",
	}

	database.ApplyRuntimeEnvironment()

	expected := map[string]string{
		"ZBX_DB_SOCKET":   "/run/mysqld/mysqld.sock",
		"ZBX_DB_HOST":     "mysql-server",
		"ZBX_DB_PORT":     "3306",
		"ZBX_DB_NAME":     "zabbix",
		"ZBX_DB_USER":     "zabbix",
		"ZBX_DB_PASSWORD": "secret",
	}
	for name, value := range expected {
		if env[name] != value {
			t.Fatalf("%s = %q, want %q", name, env[name], value)
		}
	}
}

func TestVerifiedTLSConfigurations(t *testing.T) {
	for _, test := range []struct {
		mode             string
		wantServerName   string
		wantCustomVerify bool
	}{
		{mode: "verify_ca", wantCustomVerify: true},
		{mode: "verify_full", wantServerName: "database.example.test"},
	} {
		t.Run(test.mode, func(t *testing.T) {
			env := bootstrap.Environment{
				"DB_SERVER_HOST":   "database.example.test",
				"ZBX_DBTLSCONNECT": test.mode,
			}
			database := New(env)
			config, err := database.tlsConfig()
			if err != nil {
				t.Fatal(err)
			}
			if config.ServerName != test.wantServerName {
				t.Fatalf("ServerName = %q, want %q", config.ServerName, test.wantServerName)
			}
			if (config.VerifyConnection != nil) != test.wantCustomVerify {
				t.Fatalf("VerifyConnection configured = %t", config.VerifyConnection != nil)
			}
		})
	}
}

func TestTLSClientCertificateRequiresBothFiles(t *testing.T) {
	env := bootstrap.Environment{
		"ZBX_DBTLSCONNECT":  "required",
		"ZBX_DBTLSCERTFILE": "/certificate.pem",
	}
	database := New(env)
	if _, err := database.tlsConfig(); err == nil {
		t.Fatal("incomplete client certificate configuration was accepted")
	}
}

func TestExistingSchemaIsNotImported(t *testing.T) {
	env := bootstrap.Environment{}
	database := New(env)
	database.name = "zabbix"
	admin := &fakeDatabaseSession{queries: map[string]string{
		"SELECT 1 FROM information_schema.tables WHERE table_schema = ? AND table_name = 'dbversion'": "1",
		"SELECT mandatory FROM `zabbix`.dbversion":                                                    "8000000",
	}}
	if err := database.createSchema(admin, "/does/not/exist.sql.gz"); err != nil {
		t.Fatal(err)
	}
	if _, found := env["ZBX_DB_VERSION"]; found {
		t.Fatal("database version was exposed through the environment")
	}
}

func TestAccessors(t *testing.T) {
	database := &Database{name: "zabbix", zabbixUser: "zabbix", zabbixPassword: "password"}
	if database.Name() != "zabbix" || database.User() != "zabbix" || database.Password() != "password" {
		t.Fatal("database accessors returned unexpected values")
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

func importedStatements(sessions []*fakeDatabaseSession) string {
	var statements []string
	for _, s := range sessions {
		statements = append(statements, s.statements...)
	}
	return strings.Join(statements, "\n")
}
