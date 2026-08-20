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

	path := filepath.Join(directory, "ZBX_TLSPSKFILE")
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "secret" {
		t.Fatalf("TLS file: %q, %v", data, err)
	}
	if env["ZBX_TLSPSKFILE"] != path {
		t.Fatalf("unexpected TLS file path: %q", env["ZBX_TLSPSKFILE"])
	}
	if _, found := env["ZBX_TLSPSK"]; found {
		t.Fatal("ZBX_TLSPSK was not removed")
	}

	ClearPrivateEnv(env, "ZBX_")
	if _, found := env["ZBX_TLSPSKFILE"]; found {
		t.Fatal("ZBX_TLSPSKFILE was not removed")
	}
	if env["UNRELATED_VARIABLE"] != "value" {
		t.Fatal("unrelated variable was unexpectedly removed")
	}
	if !strings.Contains(strings.Join(env.List(), "\n"), "VALUE=a=b") {
		t.Fatalf("environment list: %q", env.List())
	}
}

func TestProcessTLSFilesPreservesAbsolutePaths(t *testing.T) {
	homeDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(homeDir, "enc_internal"), 0o700); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(t.TempDir(), "agent.conf")
	if err := os.WriteFile(configPath, []byte("# TLSCAFile=\n# TLSPSKFile=\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	absoluteCA := filepath.Join(t.TempDir(), "ca.pem")
	env := Environment{
		"ZBX_TLSCAFILE": absoluteCA,
		"ZBX_TLSPSK":    "secret",
	}
	if err := ProcessTLSFiles(env, homeDir, configPath, "TLSCAFile", "TLSPSKFile"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "TLSCAFile="+absoluteCA+"\n") {
		t.Fatalf("absolute TLS path was changed:\n%s", data)
	}
	if !strings.Contains(content, "TLSPSKFile="+filepath.Join(homeDir, "enc_internal", "TLSPSKFile")+"\n") {
		t.Fatalf("inline TLS value was not written to the internal path:\n%s", data)
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

func TestRequiredDirectory(t *testing.T) {
	homeDir := t.TempDir()
	got, err := requiredDirectory(Environment{"ZABBIX_USER_HOME_DIR": homeDir}, "ZABBIX_USER_HOME_DIR")
	if err != nil {
		t.Fatal(err)
	}
	if got != homeDir {
		t.Fatalf("requiredDirectory() = %q, want %q", got, homeDir)
	}
}

func TestRequiredDirectoryRejectsInvalidPaths(t *testing.T) {
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
			if _, err := requiredDirectory(Environment{"ZABBIX_USER_HOME_DIR": test.path}, "ZABBIX_USER_HOME_DIR"); err == nil {
				t.Fatal("requiredDirectory() unexpectedly succeeded")
			}
		})
	}
}
