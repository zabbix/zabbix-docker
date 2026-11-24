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
	t.Setenv("PATH", t.TempDir())

	root := t.TempDir()
	configDir := filepath.Join(root, "etc")
	homeDir := filepath.Join(root, "home")
	pluginDirectory := filepath.Join(configDir, "zabbix_agent2.d", "plugins.d")
	if err := os.MkdirAll(filepath.Join(homeDir, "enc_internal"), 0o700); err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string]string{
		"zabbix_agent2_item_keys.conf":       "# DenyKey=system.run[*]\n",
		"zabbix_agent2_aliases.conf":         "# Alias=\n",
		"zabbix_agent2_user_parameters.conf": "# UserParameter=\n",
	} {
		testutil.WriteFile(t, filepath.Join(configDir, name), content)
	}
	for _, name := range []string{"mongodb.conf", "postgresql.conf", "mssql.conf", "ember.conf", "ceph.conf"} {
		testutil.WriteFile(t, filepath.Join(pluginDirectory, name), "# plugin config\n")
	}

	env := bootstrap.Environment{
		"ZABBIX_CONF_DIR": configDir, "ZABBIX_USER_HOME_DIR": homeDir,
		"ZBX_SERVER_HOST": "server", "ZBX_ENABLEPERSISTENTBUFFER": "true",
		"ZBX_PERSISTENTBUFFERFILE": "/buffer/agent2.db", "ZBX_ENABLESTATUSPORT": "true",
		"ZBX_ALLOWKEY_0":       "system.localtime",
		"ZBX_DENYKEY_REGEXP_1": `^system\.run\[.*\]$`,
		"ZBX_ALIAS_0":          "custom.echo:system.localtime",
		"ZBX_USERPARAMETER_0":  `custom.echo[*],printf '%s,%s' "$1" "$2"`,
		"ZBX_TLSPSK":           "secret", "POSTGRES_PASSWORD": "password",
	}
	if err := prepareService(env); err != nil {
		t.Fatal(err)
	}
	if env["ZBX_ENABLEPERSISTENTBUFFER"] != "1" || env["ZBX_STATUSPORT"] != "31999" {
		t.Fatalf("unexpected feature switches: %#v", env)
	}
	if env["POSTGRES_PASSWORD"] != "password" {
		t.Fatal("POSTGRES_PASSWORD was unexpectedly removed")
	}
	if _, found := env["ZABBIX_CONF_DIR"]; found {
		t.Fatal("ZABBIX_CONF_DIR was not removed")
	}
	mongodb := testutil.ReadFile(t, filepath.Join(pluginDirectory, "mongodb.conf"))
	if !strings.Contains(mongodb, "Plugins.MongoDB.System.Path=/usr/sbin/zabbix-agent2-plugin/mongodb") {
		t.Fatalf("MongoDB plugin config: %s", mongodb)
	}
	ceph := testutil.ReadFile(t, filepath.Join(pluginDirectory, "ceph.conf"))
	if !strings.Contains(ceph, "Plugins.Ceph.System.Path=/usr/sbin/zabbix-agent2-plugin/ceph") {
		t.Fatalf("Ceph plugin config: %s", ceph)
	}
	itemKeys := testutil.ReadFile(t, filepath.Join(configDir, "zabbix_agent2_item_keys.conf"))
	if !strings.HasSuffix(itemKeys, "AllowKey=${ZBX_ALLOWKEY_0}\nDenyKeyRegexp=${ZBX_DENYKEY_REGEXP_1}\n") {
		t.Fatalf("item key config: %s", itemKeys)
	}
	aliases := testutil.ReadFile(t, filepath.Join(configDir, "zabbix_agent2_aliases.conf"))
	if !strings.Contains(aliases, "Alias=${ZBX_ALIAS_0}") {
		t.Fatalf("alias config: %s", aliases)
	}
	userParameters := testutil.ReadFile(t, filepath.Join(configDir, "zabbix_agent2_user_parameters.conf"))
	if !strings.Contains(userParameters, "UserParameter=${ZBX_USERPARAMETER_0}") {
		t.Fatalf("user parameter config: %s", userParameters)
	}
	if env["ZBX_ALIAS_0"] == "" || env["ZBX_USERPARAMETER_0"] == "" {
		t.Fatal("indexed configuration environment variable was removed")
	}
}

func TestDisabledPersistentBuffer(t *testing.T) {
	env := bootstrap.Environment{
		"ZBX_ENABLEPERSISTENTBUFFER": "TRUE",
		"ZBX_PERSISTENTBUFFERFILE":   "/buffer/agent2.db",
		"ZBX_ENABLESTATUSPORT":       "false",
		"ZBX_STATUSPORT":             "31999",
	}
	configureFeatureSwitches(env)

	if _, found := env["ZBX_PERSISTENTBUFFERFILE"]; found {
		t.Fatal("persistent buffer file was retained")
	}
	if _, found := env["ZBX_ENABLEPERSISTENTBUFFER"]; found {
		t.Fatal("persistent buffer was enabled by a case-insensitive value")
	}
	if _, found := env["ZBX_STATUSPORT"]; found {
		t.Fatal("status port was retained")
	}
}
