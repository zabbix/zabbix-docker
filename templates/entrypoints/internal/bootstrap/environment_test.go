package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnvironmentDefaults(t *testing.T) {
	env := Environment{"EMPTY": "", "VALUE": "configured"}
	if got := env.ValueOrDefault("EMPTY", "default"); got != "" {
		t.Fatalf("ValueOrDefault() = %q, want empty value", got)
	}
	if got := env.ValueOrDefaultNonEmpty("EMPTY", "default"); got != "default" {
		t.Fatalf("ValueOrDefaultNonEmpty() = %q, want default", got)
	}
	if got := env.ValueOrDefaultNonEmpty("VALUE", "default"); got != "configured" {
		t.Fatalf("ValueOrDefaultNonEmpty() = %q, want configured", got)
	}
}

func TestProcessFileAndClearEnvironment(t *testing.T) {
	directory := t.TempDir()
	env := Environment{
		"ZBX_TLSPSK": "secret", "ZABBIX_CONF_DIR": "/etc/zabbix",
		"MYSQL_PASSWORD": "password", "VALUE": "a=b",
	}
	if err := processFileFromEnvironment(env, directory, "ZBX_TLSPSK"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(directory, "ZBX_TLSPSKFILE"))
	if err != nil || string(data) != "secret" {
		t.Fatalf("TLS file: %q, %v", data, err)
	}
	if env["ZBX_TLSPSKFILE"] != filepath.Join(directory, "ZBX_TLSPSKFILE") {
		t.Fatalf("unexpected TLS file path: %q", env["ZBX_TLSPSKFILE"])
	}
	if _, found := env["ZBX_TLSPSK"]; found {
		t.Fatal("ZBX_TLSPSK was not removed")
	}

	ClearPrivateEnv(env)
	if _, found := env["MYSQL_PASSWORD"]; found {
		t.Fatal("MYSQL_PASSWORD was not removed")
	}
	if !strings.Contains(strings.Join(env.List(), "\n"), "VALUE=a=b") {
		t.Fatalf("environment list: %q", env.List())
	}
}

func TestClearPrivateEnvWithPrefixes(t *testing.T) {
	env := Environment{
		"ZABBIX_CONF_DIR": "/etc/zabbix",
		"MYSQL_PASSWORD":  "password",
	}
	ClearPrivateEnv(env, "ZABBIX_")
	if _, found := env["ZABBIX_CONF_DIR"]; found {
		t.Fatal("ZABBIX_CONF_DIR was not removed")
	}
	if env["MYSQL_PASSWORD"] != "password" {
		t.Fatal("MYSQL_PASSWORD was unexpectedly removed")
	}
}

func TestFileEnv(t *testing.T) {
	path := filepath.Join(t.TempDir(), "password")
	if err := os.WriteFile(path, []byte("secret\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	env := Environment{"MYSQL_PASSWORD_FILE": path}
	if err := FileEnv(env, "MYSQL_PASSWORD", ""); err != nil {
		t.Fatal(err)
	}
	if env["MYSQL_PASSWORD"] != "secret" {
		t.Fatalf("MYSQL_PASSWORD = %q", env["MYSQL_PASSWORD"])
	}
	if _, found := env["MYSQL_PASSWORD_FILE"]; found {
		t.Fatal("MYSQL_PASSWORD_FILE was not removed")
	}
}

func TestRequiredHomeDirectory(t *testing.T) {
	homeDir := t.TempDir()
	got, err := RequiredHomeDirectory(Environment{"ZABBIX_USER_HOME_DIR": homeDir})
	if err != nil {
		t.Fatal(err)
	}
	if got != homeDir {
		t.Fatalf("RequiredHomeDirectory() = %q, want %q", got, homeDir)
	}
}

func TestRequiredHomeDirectoryRejectsInvalidPaths(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "home")
	if err := os.WriteFile(filePath, nil, 0o600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		path string
	}{
		{name: "missing"},
		{name: "not found", path: filepath.Join(t.TempDir(), "missing")},
		{name: "regular file", path: filePath},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := RequiredHomeDirectory(Environment{"ZABBIX_USER_HOME_DIR": test.path}); err == nil {
				t.Fatal("RequiredHomeDirectory() unexpectedly succeeded")
			}
		})
	}
}
