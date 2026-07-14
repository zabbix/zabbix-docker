//go:build !windows

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

func TestPrepareService(t *testing.T) {
	root := t.TempDir()
	configDirectory := filepath.Join(root, "etc")
	homeDirectory := filepath.Join(root, "home")
	if err := os.MkdirAll(filepath.Join(homeDirectory, "enc_internal"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(configDirectory, 0o700); err != nil {
		t.Fatal(err)
	}
	hooksDirectory := filepath.Join(homeDirectory, "entrypoint.d")
	if err := os.Mkdir(hooksDirectory, 0o700); err != nil {
		t.Fatal(err)
	}
	hookOutput := filepath.Join(root, "hook-output")
	if err := os.WriteFile(
		filepath.Join(hooksDirectory, "10-environment.sh"),
		[]byte("printf '%s' \"$ZABBIX_CONF_DIR\" > \"$HOOK_OUTPUT\"\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string]string{
		"zabbix_agentd_item_keys.conf": "# DenyKey=system.run[*]\n",
		"zabbix_agentd_modules.conf":   "# LoadModule=\n",
	} {
		if err := os.WriteFile(filepath.Join(configDirectory, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	env := bootstrap.Environment{
		"ZABBIX_CONF_DIR": configDirectory, "ZABBIX_USER_HOME_DIR": homeDirectory,
		"ZBX_SERVER_HOST": "server", "ZBX_ALLOWKEY": "system.localtime",
		"ZBX_DENYKEY": "system.run[*]", "ZBX_LOADMODULE": "module.so",
		"ZBX_TLSPSK": "secret", "MYSQL_PASSWORD": "password",
		"HOOK_OUTPUT": hookOutput,
	}
	if err := prepareService(env); err != nil {
		t.Fatal(err)
	}
	if env["ZBX_PASSIVESERVERS"] != "server" || env["ZBX_ACTIVESERVERS"] != "server" {
		t.Fatalf("unexpected environment: %#v", env)
	}
	if env["MYSQL_PASSWORD"] != "password" {
		t.Fatal("MYSQL_PASSWORD was unexpectedly removed")
	}
	if _, found := env["ZABBIX_CONF_DIR"]; found {
		t.Fatal("ZABBIX_CONF_DIR was not removed")
	}
	hookData, err := os.ReadFile(hookOutput)
	if err != nil || string(hookData) != configDirectory {
		t.Fatalf("hook environment: %q, %v", hookData, err)
	}
	itemKeys, err := os.ReadFile(filepath.Join(configDirectory, "zabbix_agentd_item_keys.conf"))
	if err != nil || !strings.Contains(string(itemKeys), "AllowKey=system.localtime") {
		t.Fatalf("item key config: %s, %v", itemKeys, err)
	}
}
