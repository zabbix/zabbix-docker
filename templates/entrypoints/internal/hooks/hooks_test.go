//go:build windows

package hooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

func TestRunExecutesHooksInOrder(t *testing.T) {
	homeDir := t.TempDir()
	directory := filepath.Join(homeDir, directoryName)
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}

	output := filepath.Join(homeDir, "output")
	for name, content := range map[string]string{
		"20-second.cmd":  "@echo second:%ZABBIX_CONF_DIR%>>\"%HOOK_OUTPUT%\"\r\n",
		"10-first.cmd":   "@echo first:%ZABBIX_CONF_DIR%>>\"%HOOK_OUTPUT%\"\r\n",
		"30-ignored.txt": "@exit /b 1\r\n",
	} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	env := bootstrap.NewEnvironment(os.Environ())
	env["ZABBIX_USER_HOME_DIR"] = homeDir
	env["ZABBIX_CONF_DIR"] = `C:\zabbix\conf`
	env["HOOK_OUTPUT"] = output
	if err := Run(env); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ReplaceAll(string(data), "\r\n", "\n")
	want := "first:C:\\zabbix\\conf\nsecond:C:\\zabbix\\conf\n"
	if got != want {
		t.Fatalf("hook output = %q, want %q", got, want)
	}
}

func TestRunReturnsHookFailure(t *testing.T) {
	homeDir := t.TempDir()
	directory := filepath.Join(homeDir, directoryName)
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(directory, "10-fail.cmd")
	if err := os.WriteFile(path, []byte("@exit /b 7\r\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	env := bootstrap.NewEnvironment(os.Environ())
	env["ZABBIX_USER_HOME_DIR"] = homeDir
	err := Run(env)
	if err == nil || !strings.Contains(err.Error(), "10-fail.cmd") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunIgnoresMissingDirectory(t *testing.T) {
	err := Run(bootstrap.Environment{"ZABBIX_USER_HOME_DIR": t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
}
