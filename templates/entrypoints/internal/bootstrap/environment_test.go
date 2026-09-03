//go:build windows

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
		"ZBX_TLSPSK": "secret", "ZABBIX_CONF_DIR": `C:\zabbix\conf`,
		"UNRELATED_VARIABLE": "value", "VALUE": "a=b",
	}
	if err := ProcessFileFromEnvironment(env, directory, "ZBX_TLSPSK"); err != nil {
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

	ClearPrivateEnv(env, "ZABBIX_")
	if env["UNRELATED_VARIABLE"] != "value" {
		t.Fatal("unrelated variable was unexpectedly removed")
	}
	if !strings.Contains(strings.Join(env.List(), "\n"), "VALUE=a=b") {
		t.Fatalf("environment list: %q", env.List())
	}
}


func TestProcessTLSFilesResolvesRelativePaths(t *testing.T) {
	homeDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(homeDir, "enc_internal"), 0o700); err != nil {
		t.Fatal(err)
	}

	absoluteCertPath := filepath.Join(homeDir, "custom", "agent.crt")
	env := Environment{
		"ZBX_TLSCAFILE":   "ca.crt",
		"ZBX_TLSCERTFILE": absoluteCertPath,
		"ZBX_TLSPSK":      "secret",
	}

	if err := ProcessTLSFiles(env, homeDir, "ZBX_TLSCA", "ZBX_TLSCERT", "ZBX_TLSPSK"); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		want string
	}{
		{
			name: "ZBX_TLSCAFILE",
			want: filepath.Join(homeDir, "enc", "ca.crt"),
		},
		{
			name: "ZBX_TLSCERTFILE",
			want: absoluteCertPath,
		},
		{
			name: "ZBX_TLSPSKFILE",
			want: filepath.Join(homeDir, "enc_internal", "ZBX_TLSPSKFILE"),
		},
	}
	for _, test := range tests {
		if got := env[test.name]; got != test.want {
			t.Errorf("%s = %q, want %q", test.name, got, test.want)
		}
	}

	if _, found := env["ZBX_TLSPSK"]; found {
		t.Error("plaintext ZBX_TLSPSK was not removed from the environment")
	}

	data, err := os.ReadFile(env["ZBX_TLSPSKFILE"])
	if err != nil {
		t.Fatal(err)
	}
	if got := string(data); got != "secret" {
		t.Errorf("ZBX_TLSPSKFILE content = %q, want %q", got, "secret")
	}
}

func TestClearPrivateEnvWithPrefixes(t *testing.T) {
	env := Environment{
		"ZABBIX_CONF_DIR":    `C:\zabbix\conf`,
		"UNRELATED_VARIABLE": "value",
	}
	ClearPrivateEnv(env, "ZABBIX_")
	if _, found := env["ZABBIX_CONF_DIR"]; found {
		t.Fatal("ZABBIX_CONF_DIR was not removed")
	}
	if env["UNRELATED_VARIABLE"] != "value" {
		t.Fatal("unrelated variable was unexpectedly removed")
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
