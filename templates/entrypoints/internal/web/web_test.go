package web

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
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
	if err := preparePHP(env, MySQL); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"ZBX_HISTORYPROVIDERS":       "[]",
		"ZBX_CERT_STORAGE":           "database",
		"ZBX_BANNERS_ENABLED":        "true",
		"ZBX_HTTP_AUTH_ENABLED":      "true",
		"ZBX_MODULES_CONFIG_ENABLED": "true",
		"ZBX_MEDIA_TYPE_DENYLIST":    "[]",
		"ZBX_DB_DOUBLE_IEEE754":      "true",
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

func TestClearWebEnv(t *testing.T) {
	tests := []struct {
		name            string
		env             bootstrap.Environment
		wantCredentials bool
		wantPrivate     bool
	}{
		{
			name:            "direct credentials",
			env:             bootstrap.Environment{},
			wantCredentials: true,
		},
		{
			name: "HashiCorp Vault",
			env: bootstrap.Environment{
				"ZBX_VAULT": "HashiCorp", "VAULT_TOKEN": "token", "ZBX_VAULTURL": "https://vault",
			},
		},
		{
			name: "HashiCorp Vault without token",
			env: bootstrap.Environment{
				"ZBX_VAULT": "HashiCorp", "ZBX_VAULTURL": "https://vault",
			},
		},
		{
			name: "CyberArk Vault",
			env:  bootstrap.Environment{"ZBX_VAULT": "CyberArk"},
		},
		{
			name: "cleanup disabled",
			env: bootstrap.Environment{
				"ZBX_VAULT": "HashiCorp", "VAULT_TOKEN": "token", "ZBX_VAULTURL": "https://vault",
				"ZBX_CLEAR_ENV": "false",
			},
			wantCredentials: true,
			wantPrivate:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.env["DB_SERVER_USER"] = "zabbix"
			test.env["DB_SERVER_PASS"] = "secret"
			test.env["MYSQL_PASSWORD"] = "mysql-secret"
			test.env["NGINX_TEST"] = "private"

			clearWebEnv(test.env)

			_, found := test.env["MYSQL_PASSWORD"]
			if found != test.wantPrivate {
				t.Fatalf("MYSQL_PASSWORD presence = %t, want %t", found, test.wantPrivate)
			}
			if test.env["NGINX_TEST"] != "private" {
				t.Fatal("NGINX_TEST was unexpectedly removed")
			}
			for _, name := range []string{"DB_SERVER_USER", "DB_SERVER_PASS"} {
				_, found := test.env[name]
				if found != test.wantCredentials {
					t.Fatalf("%s presence = %t, want %t", name, found, test.wantCredentials)
				}
			}
		})
	}
}

func TestWebServerEnvRemovesDatabaseCredentials(t *testing.T) {
	env := bootstrap.Environment{
		"DB_SERVER_USER":    "zabbix",
		"DB_SERVER_PASS":    "db-secret",
		"MYSQL_PASSWORD":    "mysql-secret",
		"POSTGRES_PASSWORD": "postgres-secret",
		"ZBX_DB_PASSWORD":   "frontend-secret",
		"ZBX_VAULT":         "HashiCorp",
		"ZBX_VAULTDBPATH":   "secret/zabbix",
		"VAULT_TOKEN":       "vault-secret",
		"NGINX_BIN":         "/custom/nginx",
		"PATH":              "/usr/bin",
	}

	webEnv := webServerEnv(env)
	for _, name := range []string{
		"DB_SERVER_USER", "DB_SERVER_PASS", "MYSQL_PASSWORD", "POSTGRES_PASSWORD",
		"ZBX_DB_PASSWORD", "ZBX_VAULT", "ZBX_VAULTDBPATH", "VAULT_TOKEN",
	} {
		if _, found := webEnv[name]; found {
			t.Fatalf("%s was passed to the web server", name)
		}
	}
	for _, name := range []string{"NGINX_BIN", "PATH"} {
		if webEnv[name] != env[name] {
			t.Fatalf("%s = %q, want %q", name, webEnv[name], env[name])
		}
	}
	if env["DB_SERVER_PASS"] != "db-secret" || env["VAULT_TOKEN"] != "vault-secret" {
		t.Fatal("source PHP-FPM environment was modified")
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
	err := startStack(env, Nginx)
	var exitError *exec.ExitError
	if !errors.As(err, &exitError) {
		t.Fatalf("startStack() error = %T, want *exec.ExitError: %v", err, err)
	}
	if code := exitError.ExitCode(); code != 23 {
		t.Fatalf("startStack() exit code = %d, want 23: %v", code, err)
	}
}
