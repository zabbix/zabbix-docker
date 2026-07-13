package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

func TestPrepareServiceWindows(t *testing.T) {
	root := t.TempDir()
	configDirectory := filepath.Join(root, "conf")
	homeDirectory := filepath.Join(root, "home")
	if err := os.MkdirAll(filepath.Join(homeDirectory, "enc_internal"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(configDirectory, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDirectory, "zabbix_agentd_item_keys.conf"), []byte("# DenyKey=system.run[*]\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	env := bootstrap.Environment{
		"ZABBIX_CONF_DIR": configDirectory, "ZABBIX_USER_HOME_DIR": homeDirectory,
		"MYSQL_PASSWORD": "password",
	}
	if err := prepareService(env); err != nil {
		t.Fatal(err)
	}
	if env["ZBX_PASSIVESERVERS"] != "zabbix-server" || env["ZBX_ACTIVESERVERS"] != "zabbix-server" {
		t.Fatalf("unexpected server configuration: %#v", env)
	}
	if env["MYSQL_PASSWORD"] != "password" {
		t.Fatal("Windows entrypoint removed a non-ZABBIX variable")
	}
	if _, found := env["ZABBIX_CONF_DIR"]; found {
		t.Fatal("ZABBIX_CONF_DIR was not removed")
	}
}
