package web

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

func TestPreparePHPUsesTrunkFrontendSettings(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "enc_internal"), 0o700); err != nil {
		t.Fatal(err)
	}
	phpConfig := filepath.Join(root, "php-fpm.conf")
	if err := os.WriteFile(phpConfig, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	env := bootstrap.Environment{
		"ZABBIX_USER_HOME_DIR": root,
		"PHP_ZBX_CONFIG_FILE":  phpConfig,
		"DAEMON_USER":          "zabbix",
		"DAEMON_GROUP":         "zabbix",
	}
	if err := preparePHP(env, DatabaseMySQL); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"ZBX_HISTORYPROVIDERS":       "[]",
		"ZBX_CERT_STORAGE":           "database",
		"ZBX_BANNERS_ENABLED":        "true",
		"ZBX_HTTP_AUTH_ENABLED":      "true",
		"ZBX_MODULES_CONFIG_ENABLED": "true",
		"ZBX_MEDIA_TYPE_DENYLIST":    "[]",
	}
	for name, value := range want {
		if env[name] != value {
			t.Fatalf("%s = %q, want %q", name, env[name], value)
		}
	}
	if value, found := env["DB_SERVER_SOCKET"]; !found || value != "" {
		t.Fatalf("DB_SERVER_SOCKET = %q, found = %t", value, found)
	}
	for _, optional := range []string{"ZBX_VAULT", "ZBX_SSO_SETTINGS", "ZBX_SERVER_TLS_CERT_ISSUER"} {
		if _, found := env[optional]; found {
			t.Fatalf("unset optional variable %s was exported", optional)
		}
	}
	for _, obsolete := range []string{"ZBX_HISTORYSTORAGEURL", "ZBX_HISTORYSTORAGETYPES", "ZBX_ALLOW_HTTP_AUTH"} {
		if _, found := env[obsolete]; found {
			t.Fatalf("obsolete frontend setting %s was exported", obsolete)
		}
	}
}

func TestPrepareSessionName(t *testing.T) {
	root := t.TempDir()
	include := filepath.Join(root, "include")
	if err := os.Mkdir(include, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(include, "defines.inc.php")
	if err := os.WriteFile(path, []byte("define('ZBX_SESSION_NAME', 'zbx_sessionid');\n"), 0o640); err != nil {
		t.Fatal(err)
	}
	env := bootstrap.Environment{"ZABBIX_WWW_ROOT": root, "ZBX_SESSION_NAME": "custom$name"}
	if err := prepareSessionName(env); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "'custom$name'") {
		t.Fatalf("session name was not replaced: %s", data)
	}
}

func TestPrepareNginx(t *testing.T) {
	root := t.TempDir()
	includes := filepath.Join(root, "includes")
	confDirectory := filepath.Join(root, "conf.d")
	zabbixConfig := filepath.Join(root, "zabbix")
	for _, directory := range []string{includes, confDirectory, zabbixConfig} {
		if err := os.Mkdir(directory, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	serverCommon := filepath.Join(includes, "server-common.conf")
	nginxConfig := filepath.Join(root, "nginx.conf")
	if err := os.WriteFile(serverCommon, []byte("timeout {FCGI_READ_TIMEOUT}; index {HTTP_INDEX_FILE};\n"), 0o640); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(nginxConfig, []byte("server_tokens {EXPOSE_WEB_SERVER_INFO};\n"), 0o640); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(zabbixConfig, "nginx.conf"), []byte("http"), 0o640); err != nil {
		t.Fatal(err)
	}
	env := bootstrap.Environment{
		"NGINX_INCLUDES_DIR": includes, "NGINX_CONFD_DIR": confDirectory, "NGINX_CONF_FILE": nginxConfig,
		"ZABBIX_CONF_DIR": zabbixConfig, "ZBX_MAXEXECUTIONTIME": "600", "HTTP_INDEX_FILE": "index.php",
		"EXPOSE_WEB_SERVER_INFO": "off", "ENABLE_WEB_ACCESS_LOG": "false", "WEB_REAL_IP_FROM": "10.0.0.0/8", "WEB_REAL_IP_HEADER": "X-Forwarded-For",
	}
	if err := prepareNginx(env); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(serverCommon)
	if string(data) != "timeout 601; index index.php;\n" {
		t.Fatalf("unexpected server config: %s", data)
	}
	data, _ = os.ReadFile(filepath.Join(includes, "real-ip.conf"))
	if string(data) != "set_real_ip_from 10.0.0.0/8;\nreal_ip_header X-Forwarded-For;\n" {
		t.Fatalf("unexpected real IP config: %s", data)
	}
}

func TestPrepareNginxReturnsSymlinkErrors(t *testing.T) {
	root := t.TempDir()
	includes := filepath.Join(root, "includes")
	zabbixConfig := filepath.Join(root, "zabbix")
	for _, directory := range []string{includes, zabbixConfig} {
		if err := os.Mkdir(directory, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(zabbixConfig, "nginx.conf"), []byte("http"), 0o640); err != nil {
		t.Fatal(err)
	}

	env := bootstrap.Environment{
		"NGINX_INCLUDES_DIR": includes,
		"NGINX_CONFD_DIR":    filepath.Join(root, "missing", "conf.d"),
		"ZABBIX_CONF_DIR":    zabbixConfig,
	}
	if err := prepareNginx(env); err == nil {
		t.Fatal("filesystem error while enabling the virtual host was ignored")
	}
}

func TestStartStackReturnsFirstProcessExitCode(t *testing.T) {
	root := t.TempDir()
	phpBinary := filepath.Join(root, "php-fpm")
	nginxBinary := filepath.Join(root, "nginx")
	if err := os.WriteFile(phpBinary, []byte("#!/bin/sh\nexit 23\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(nginxBinary, []byte("#!/bin/sh\ntrap 'exit 0' TERM INT\nwhile :; do /bin/sleep 1; done\n"), 0o700); err != nil {
		t.Fatal(err)
	}

	env := bootstrap.Environment{
		"PHP_FPM_BIN": phpBinary,
		"NGINX_BIN":   nginxBinary,
	}
	err := startStack(env, ServerNginx)
	if code := bootstrap.ExitCode(err); code != 23 {
		t.Fatalf("startStack() exit code = %d, want 23: %v", code, err)
	}
}
