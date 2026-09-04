//go:build windows

package config

import (
	"testing"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

func TestConfigureServers(t *testing.T) {
	env := bootstrap.Environment{
		"ZBX_SERVER_HOST": "server", "ZBX_SERVER_PORT": "10061",
		"ZBX_PASSIVESERVERS": "passive", "ZBX_ACTIVESERVERS": "active",
	}
	ConfigureServers(env)
	if env["ZBX_PASSIVESERVERS"] != "server,passive" || env["ZBX_ACTIVESERVERS"] != "server:10061,active" {
		t.Fatalf("unexpected servers: %#v", env)
	}
}
