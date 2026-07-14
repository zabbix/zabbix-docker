//go:build !windows

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
		"20-second.sh": "printf 'second:%s\\n' \"$ZABBIX_CONF_DIR\" >> \"$HOOK_OUTPUT\"\n",
		"10-first.sh":  "printf 'first:%s\\n' \"$ZABBIX_CONF_DIR\" >> \"$HOOK_OUTPUT\"\n",
		"30-ignored":   "exit 1\n",
	} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	env := bootstrap.Environment{
		"ZABBIX_USER_HOME_DIR": homeDir,
		"ZABBIX_CONF_DIR":      "/etc/zabbix",
		"HOOK_OUTPUT":          output,
	}
	if err := Run(env); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), "first:/etc/zabbix\nsecond:/etc/zabbix\n"; got != want {
		t.Fatalf("hook output = %q, want %q", got, want)
	}
}

func TestRunReturnsHookFailure(t *testing.T) {
	homeDir := t.TempDir()
	directory := filepath.Join(homeDir, directoryName)
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(directory, "10-fail.sh")
	if err := os.WriteFile(path, []byte("exit 7\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	err := Run(bootstrap.Environment{"ZABBIX_USER_HOME_DIR": homeDir})
	if err == nil || !strings.Contains(err.Error(), "10-fail.sh") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunIgnoresMissingDirectory(t *testing.T) {
	err := Run(bootstrap.Environment{"ZABBIX_USER_HOME_DIR": t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
}
