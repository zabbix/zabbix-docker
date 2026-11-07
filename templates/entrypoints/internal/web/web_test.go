package web

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

func TestPreparePHPUsesFrontendSettings(t *testing.T) {
	root := t.TempDir()
	phpConfig := filepath.Join(root, "php-fpm.conf")
	if err := os.WriteFile(phpConfig, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	env := bootstrap.Environment{
		"ZABBIX_USER_HOME_DIR": root,
		"PHP_ZBX_CONFIG_FILE":  phpConfig,
		"DAEMON_USER":          "zabbix",
		"DAEMON_GROUP":         "zabbix",
		"ZBX_DB_CIPHER_LIST":   "cipher",
	}
	if err := prepareFrontendAliases(env); err != nil {
		t.Fatal(err)
	}
	if env["ZBX_DB_VERIFY_HOST"] != "true" {
		t.Fatal("database hostname verification must default to true before the bootstrap connection")
	}
	if err := preparePHP(env, MySQL); err != nil {
		t.Fatal(err)
	}
	if value, found := env["ZBX_SERVER_HOST"]; found {
		t.Fatalf("ZBX_SERVER_HOST got default %q; unset must keep the address from the database", value)
	}
	want := map[string]string{
		"ZBX_DB_VERIFY_HOST":         "true",
		"ZBX_MAXEXECUTIONTIME":       "300",
		"ZBX_HISTORYPROVIDERS":       "[]",
		"ZBX_TELEMETRYPROVIDERS":     "[]",
		"ZBX_CERT_STORAGE":           "database",
		"ZBX_BANNERS_ENABLED":        "true",
		"ZBX_HTTP_AUTH_ENABLED":      "true",
		"ZBX_MODULES_CONFIG_ENABLED": "true",
		"ZBX_MEDIA_TYPE_DENYLIST":    "[]",
		"ZBX_DB_DOUBLE_IEEE754":      "true",
		"ZBX_DBTLSCIPHER":            "cipher",
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
	for _, legacy := range []string{"ZBX_DB_CIPHER_LIST"} {
		if _, found := env[legacy]; found {
			t.Fatalf("legacy variable %s was exported", legacy)
		}
	}
	explicitEnv := bootstrap.Environment{
		"ZABBIX_USER_HOME_DIR": root,
		"PHP_ZBX_CONFIG_FILE":  phpConfig,
		"DAEMON_USER":          "zabbix",
		"DAEMON_GROUP":         "zabbix",
		"ZBX_DB_VERIFY_HOST":   "false",
	}
	if err := prepareFrontendAliases(explicitEnv); err != nil {
		t.Fatal(err)
	}
	if err := preparePHP(explicitEnv, MySQL); err != nil {
		t.Fatal(err)
	}
	if explicitEnv["ZBX_DB_VERIFY_HOST"] != "false" {
		t.Fatal("explicitly disabled database hostname verification was overridden")
	}
}

func TestPrepareFrontendAliasesCipherPriority(t *testing.T) {
	env := bootstrap.Environment{
		"ZBX_DB_CIPHER_LIST": "legacy",
		"ZBX_DBTLSCIPHER":    "current",
	}
	if err := prepareFrontendAliases(env); err != nil {
		t.Fatal(err)
	}
	if env["ZBX_DBTLSCIPHER"] != "current" {
		t.Fatalf("ZBX_DBTLSCIPHER=%q, want current", env["ZBX_DBTLSCIPHER"])
	}
	if _, found := env["ZBX_DB_CIPHER_LIST"]; found {
		t.Fatal("ZBX_DB_CIPHER_LIST was not removed")
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
			name: "HashiCorp Vault for a non-database secret",
			env: bootstrap.Environment{
				"ZBX_VAULT": "HashiCorp", "VAULT_TOKEN": "token", "ZBX_VAULTURL": "https://vault",
			},
			wantCredentials: true,
		},
		{
			name: "HashiCorp Vault database credentials",
			env: bootstrap.Environment{
				"ZBX_VAULT": "HashiCorp", "ZBX_VAULTURL": "https://vault", "ZBX_VAULTDBPATH": "secret/database",
			},
		},
		{
			name:            "CyberArk Vault for a non-database secret",
			env:             bootstrap.Environment{"ZBX_VAULT": "CyberArk"},
			wantCredentials: true,
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
		"DB_SERVER_USER":                           "zabbix",
		"DB_SERVER_PASS":                           "db-secret",
		"MYSQL_PASSWORD":                           "mysql-secret",
		"POSTGRES_PASSWORD":                        "postgres-secret",
		"ZBX_DB_PASSWORD":                          "frontend-secret",
		"ZBX_DBTLSKEYFILE":                         "/run/secrets/client-key.pem",
		"ZBX_VAULT":                                "HashiCorp",
		"ZBX_VAULTDBPATH":                          "secret/zabbix",
		"VAULT_TOKEN":                              "vault-secret",
		"ZBX_TELEMETRYPROVIDERS":                   `[{"provider":"clickhouse","password":"telemetry-secret"}]`,
		"ZBX_TELEMETRYPROVIDER_0_PASSWORD":         "telemetry-secret",
		"ZBX_TELEMETRYPROVIDER_0_PASSWORD_FILE":    "/run/secrets/telemetry-password",
		"ZBX_TELEMETRYPROVIDER_0_SSL_KEY_PASSWORD": "telemetry-key-secret",
		"NGINX_BIN":                                "/custom/nginx",
		"PATH":                                     "/usr/bin",
	}

	webEnv := webServerEnv(env)
	for _, name := range []string{
		"DB_SERVER_USER", "DB_SERVER_PASS", "MYSQL_PASSWORD", "POSTGRES_PASSWORD",
		"ZBX_DB_PASSWORD", "ZBX_DBTLSKEYFILE", "ZBX_VAULT", "ZBX_VAULTDBPATH", "VAULT_TOKEN",
		"ZBX_TELEMETRYPROVIDERS", "ZBX_TELEMETRYPROVIDER_0_PASSWORD",
		"ZBX_TELEMETRYPROVIDER_0_PASSWORD_FILE", "ZBX_TELEMETRYPROVIDER_0_SSL_KEY_PASSWORD",
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

func TestPrepareWebDisablesHTTPSWhenTLSFileIsRemoved(t *testing.T) {
	for _, tc := range []struct {
		name, enabledVHost, removedFile string
		prepare                         func(bootstrap.Environment) error
	}{
		{"nginx", "nginx", "dhparam.pem", prepareNginx},
		{"apache", "zabbix", "ssl.key", prepareApache},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			configDir := filepath.Join(root, "zabbix")
			sitesDir := filepath.Join(root, "sites")
			sslDir := filepath.Join(root, "ssl")
			includesDir := filepath.Join(root, "includes")
			for _, dir := range []string{configDir, sitesDir, sslDir, includesDir} {
				if err := os.Mkdir(dir, 0o700); err != nil {
					t.Fatal(err)
				}
			}

			httpsSource := filepath.Join(configDir, tc.name+"_ssl.conf")
			httpsLink := filepath.Join(sitesDir, tc.enabledVHost+"_ssl.conf")
			for _, path := range []string{
				filepath.Join(configDir, tc.name+".conf"), httpsSource,
				filepath.Join(sslDir, "ssl.crt"), filepath.Join(sslDir, "ssl.key"), filepath.Join(sslDir, "dhparam.pem"),
			} {
				if err := os.WriteFile(path, nil, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			env := bootstrap.Environment{
				"ZABBIX_CONF_DIR": configDir, "ZBX_MAXEXECUTIONTIME": "600",
				"NGINX_CONF_FILE": filepath.Join(root, "nginx.conf"), "NGINX_INCLUDES_DIR": includesDir,
				"NGINX_CONFD_DIR": sitesDir, "NGINX_SSL_CONFIG_DIR": sslDir,
				"APACHE_SITES_DIR": sitesDir, "APACHE_SSL_CONFIG_DIR": sslDir, "APACHE_RUN_DIR": filepath.Join(root, "run"),
			}

			prepare := func() {
				t.Helper()
				if err := tc.prepare(env); err != nil {
					t.Fatal(err)
				}
			}
			prepare()
			if got, err := os.Readlink(httpsLink); err != nil || got != httpsSource {
				t.Fatalf("HTTPS link = %q, error = %v", got, err)
			}

			if err := os.Remove(filepath.Join(sslDir, tc.removedFile)); err != nil {
				t.Fatal(err)
			}
			prepare()
			if _, err := os.Lstat(httpsLink); !os.IsNotExist(err) {
				t.Fatalf("HTTPS link was not removed: %v", err)
			}
			if _, err := os.Readlink(filepath.Join(sitesDir, tc.enabledVHost+".conf")); err != nil {
				t.Fatalf("HTTP link is missing: %v", err)
			}

			const customTarget = "custom-https.conf"
			if err := os.Symlink(customTarget, httpsLink); err != nil {
				t.Fatal(err)
			}
			prepare()
			if got, err := os.Readlink(httpsLink); err != nil || got != customTarget {
				t.Fatalf("custom HTTPS link = %q, error = %v", got, err)
			}

			if err := os.Remove(httpsLink); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(httpsLink, []byte("custom vhost"), 0o600); err != nil {
				t.Fatal(err)
			}
			prepare()
			if got, err := os.ReadFile(httpsLink); err != nil || string(got) != "custom vhost" {
				t.Fatalf("custom HTTPS file = %q, error = %v", got, err)
			}
		})
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
