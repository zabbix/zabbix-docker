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
	configDir := filepath.Join(root, "etc")
	homeDir := filepath.Join(root, "home")
	if err := os.MkdirAll(filepath.Join(homeDir, "enc_internal"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatal(err)
	}
	hooksDirectory := filepath.Join(configDir, "entrypoint.d")
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
		"zabbix_agentd_aliases.conf":         "# Alias=\n",
		"zabbix_agentd_item_keys.conf":       "# DenyKey=system.run[*]\n",
		"zabbix_agentd_modules.conf":         "# LoadModule=\n",
		"zabbix_agentd_user_parameters.conf": "# UserParameter=\n",
	} {
		if err := os.WriteFile(filepath.Join(configDir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	env := bootstrap.Environment{
		"ZABBIX_CONF_DIR": configDir, "ZABBIX_USER_HOME_DIR": homeDir,
		"ZBX_SERVER_HOST": "server", "ZBX_LOADMODULE": "module.so",
		"ZBX_ALLOWKEY_0":      "system.localtime",
		"ZBX_DENYKEY_1":       "system.run[*]",
		"ZBX_ALIAS_0":         "custom.echo:system.localtime",
		"ZBX_USERPARAMETER_0": `custom.echo[*],printf '%s,%s' "$1" "$2"`,
		"ZBX_TLSPSK":          "secret", "MYSQL_PASSWORD": "password",
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
	if err != nil || string(hookData) != configDir {
		t.Fatalf("hook environment: %q, %v", hookData, err)
	}
	itemKeys, err := os.ReadFile(filepath.Join(configDir, "zabbix_agentd_item_keys.conf"))
	if err != nil || !strings.HasSuffix(string(itemKeys),
		"AllowKey=${ZBX_ALLOWKEY_0}\nDenyKey=${ZBX_DENYKEY_1}\n") {
		t.Fatalf("item key config: %s, %v", itemKeys, err)
	}
	aliases, err := os.ReadFile(filepath.Join(configDir, "zabbix_agentd_aliases.conf"))
	if err != nil || !strings.Contains(string(aliases), "Alias=${ZBX_ALIAS_0}") {
		t.Fatalf("alias config: %s, %v", aliases, err)
	}
	userParameters, err := os.ReadFile(filepath.Join(configDir, "zabbix_agentd_user_parameters.conf"))
	if err != nil || !strings.Contains(string(userParameters), "UserParameter=${ZBX_USERPARAMETER_0}") {
		t.Fatalf("user parameter config: %s, %v", userParameters, err)
	}
	if env["ZBX_ALIAS_0"] == "" || env["ZBX_USERPARAMETER_0"] == "" {
		t.Fatal("indexed configuration environment variable was removed")
	}
}
