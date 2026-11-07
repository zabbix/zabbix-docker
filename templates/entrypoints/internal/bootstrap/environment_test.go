package bootstrap

import (
	"os"
	"path/filepath"
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

func TestRequiredHomeDir(t *testing.T) {
	homeDir := t.TempDir()
	got, err := RequiredHomeDir(Environment{"ZABBIX_USER_HOME_DIR": homeDir})
	if err != nil {
		t.Fatal(err)
	}
	if got != homeDir {
		t.Fatalf("RequiredHomeDir() = %q, want %q", got, homeDir)
	}
}

func TestRequiredHomeDirRejectsInvalidPaths(t *testing.T) {
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
			if _, err := RequiredHomeDir(Environment{"ZABBIX_USER_HOME_DIR": test.path}); err == nil {
				t.Fatal("RequiredHomeDir() unexpectedly succeeded")
			}
		})
	}
}
