//go:build windows

package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

func TestConfigureServers(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "agent.conf")
	if err := os.WriteFile(configPath, []byte("# Server=\n# ServerActive=\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	env := bootstrap.Environment{
		"ZBX_SERVER_HOST": "server", "ZBX_SERVER_PORT": "10061",
		"ZBX_PASSIVESERVERS": "passive", "ZBX_ACTIVESERVERS": "active",
	}
	if err := ConfigureServers(env, configPath); err != nil {
		t.Fatal(err)
	}
	if env["ZBX_PASSIVESERVERS"] != "server,passive" || env["ZBX_ACTIVESERVERS"] != "server:10061,active" {
		t.Fatalf("unexpected servers: %#v", env)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "Server=server,passive\n") || !strings.Contains(content, "ServerActive=server:10061,active\n") {
		t.Fatalf("server configuration was not written:\n%s", data)
	}
}
