//go:build !windows

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/testutil"
)

func TestPrepareService(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "etc")
	homeDir := filepath.Join(root, "home")
	if err := os.MkdirAll(filepath.Join(homeDir, "enc_internal"), 0o700); err != nil {
		t.Fatal(err)
	}
	hookOutput := filepath.Join(root, "hook-output")
	testutil.WriteFile(t, filepath.Join(configDir, "entrypoint.d", "10-environment.sh"),
		"printf '%s' \"$ZABBIX_CONF_DIR\" > \"$HOOK_OUTPUT\"\n")
	for name, content := range map[string]string{
		"zabbix_agentd_aliases.conf":         "# Alias=\n",
		"zabbix_agentd_item_keys.conf":       "# DenyKey=system.run[*]\n",
		"zabbix_agentd_modules.conf":         "# LoadModule=\n",
		"zabbix_agentd_user_parameters.conf": "# UserParameter=\n",
	} {
		testutil.WriteFile(t, filepath.Join(configDir, name), content)
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
	if hook := testutil.ReadFile(t, hookOutput); hook != configDir {
		t.Fatalf("hook environment: %q", hook)
	}
	itemKeys := testutil.ReadFile(t, filepath.Join(configDir, "zabbix_agentd_item_keys.conf"))
	if !strings.HasSuffix(itemKeys, "AllowKey=${ZBX_ALLOWKEY_0}\nDenyKey=${ZBX_DENYKEY_1}\n") {
		t.Fatalf("item key config: %s", itemKeys)
	}
	aliases := testutil.ReadFile(t, filepath.Join(configDir, "zabbix_agentd_aliases.conf"))
	if !strings.Contains(aliases, "Alias=${ZBX_ALIAS_0}") {
		t.Fatalf("alias config: %s", aliases)
	}
	userParameters := testutil.ReadFile(t, filepath.Join(configDir, "zabbix_agentd_user_parameters.conf"))
	if !strings.Contains(userParameters, "UserParameter=${ZBX_USERPARAMETER_0}") {
		t.Fatalf("user parameter config: %s", userParameters)
	}
	if env["ZBX_ALIAS_0"] == "" || env["ZBX_USERPARAMETER_0"] == "" {
		t.Fatal("indexed configuration environment variable was removed")
	}
}
