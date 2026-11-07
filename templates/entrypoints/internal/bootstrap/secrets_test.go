package bootstrap

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func useSecretRoots(t *testing.T) (external, internal, standard string) {
	t.Helper()
	oldExternal, oldInternal, oldStandard := externalSecretsRoot, internalSecretsRoot, dockerSecretsRoot
	root := t.TempDir()
	external = filepath.Join(root, "external")
	internal = filepath.Join(root, "internal")
	standard = filepath.Join(root, "docker")
	for _, path := range []string{external, internal, standard} {
		if err := os.MkdirAll(path, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	externalSecretsRoot, internalSecretsRoot, dockerSecretsRoot = external, internal, standard
	t.Cleanup(func() {
		externalSecretsRoot, internalSecretsRoot, dockerSecretsRoot = oldExternal, oldInternal, oldStandard
	})
	return external, internal, standard
}

func writeTestSecret(t *testing.T, path, value string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestResolveSecretsMaterializesInline(t *testing.T) {
	_, internal, _ := useSecretRoots(t)
	env := Environment{"ZBX_TLSACCEPT": "cert", "ZBX_TLSCERT": "certificate"}
	if err := ResolveSecrets(env, SecretsMCPServer, SecretsService); err != nil {
		t.Fatal(err)
	}
	if _, exists := env["ZBX_TLSCERT"]; exists {
		t.Fatal("inline secret was not removed")
	}
	want := filepath.Join(internal, "tls", "ZBX_TLSCERTFILE")
	if env["ZBX_TLSCERTFILE"] != want {
		t.Fatalf("ZBX_TLSCERTFILE=%q, want %q", env["ZBX_TLSCERTFILE"], want)
	}
	data, err := os.ReadFile(want)
	if err != nil || string(data) != "certificate" {
		t.Fatalf("generated secret=%q, %v", data, err)
	}
	info, err := os.Stat(want)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("generated secret mode=%v", info.Mode().Perm())
	}
}

func TestResolveSecretsPrefersDockerSecrets(t *testing.T) {
	external, _, standard := useSecretRoots(t)
	writeTestSecret(t, filepath.Join(external, "database", "password"), "new")
	writeTestSecret(t, filepath.Join(standard, "password"), "standard")
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	originalStderr := os.Stderr
	os.Stderr = writer
	t.Cleanup(func() {
		os.Stderr = originalStderr
		_ = reader.Close()
		_ = writer.Close()
	})
	env := Environment{"MYSQL_PASSWORD_FILE": "password"}
	if err := ResolveSecrets(env, SecretsServerMySQL, SecretsInitDB); err != nil {
		t.Fatal(err)
	}
	os.Stderr = originalStderr
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	warning, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if env["MYSQL_PASSWORD"] != "standard" {
		t.Fatalf("MYSQL_PASSWORD=%q, want standard source", env["MYSQL_PASSWORD"])
	}
	if !strings.Contains(string(warning), "using standard Docker secret") {
		t.Fatalf("warning=%q, want standard secret selection", warning)
	}
}

func TestResolveSecretsAcceptsExplicitEmptyFile(t *testing.T) {
	_, _, standard := useSecretRoots(t)
	writeTestSecret(t, filepath.Join(standard, "password"), "")
	env := Environment{"MYSQL_PASSWORD_FILE": "password"}
	if err := ResolveSecrets(env, SecretsServerMySQL, SecretsInitDB); err != nil {
		t.Fatal(err)
	}
	if value, exists := env["MYSQL_PASSWORD"]; !exists || value != "" {
		t.Fatalf("empty secret was not preserved: value=%q, exists=%v", value, exists)
	}
	if _, exists := env["MYSQL_PASSWORD_FILE"]; exists {
		t.Fatal("MYSQL_PASSWORD_FILE was not removed")
	}
}

func TestResolveSecretsDropsExplicitlyEmptySources(t *testing.T) {
	_, _, _ = useSecretRoots(t)
	env := Environment{"MYSQL_PASSWORD": "", "MYSQL_PASSWORD_FILE": ""}
	if err := ResolveSecrets(env, SecretsServerMySQL, SecretsInitDB); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"MYSQL_PASSWORD", "MYSQL_PASSWORD_FILE"} {
		if _, exists := env[name]; exists {
			t.Fatalf("empty source %s was not removed", name)
		}
	}
}

func TestResolveSecretsRejectsNewAndDeprecatedFile(t *testing.T) {
	external, _, _ := useSecretRoots(t)
	home := t.TempDir()
	writeTestSecret(t, filepath.Join(external, "tls", "cert.pem"), "new")
	writeTestSecret(t, filepath.Join(home, "enc", "cert.pem"), "legacy")
	env := Environment{"ZABBIX_USER_HOME_DIR": home, "ZBX_TLSCERTFILE": "cert.pem"}
	err := ResolveSecrets(env, SecretsAgent, SecretsService)
	if err == nil || !strings.Contains(err.Error(), "both") {
		t.Fatalf("ResolveSecrets() error=%v, want duplicate-location error", err)
	}
}

func TestResolveSecretsDoesNotUseLegacyPathsForEightZeroSecrets(t *testing.T) {
	const provider = `{"provider":"clickhouse"}`
	for _, test := range []struct {
		name      string
		component SecretComponent
		command   SecretCommand
		env       Environment
		legacy    func(home string) string
	}{
		{
			"proxy APM",
			SecretsProxySQLite,
			SecretsService,
			Environment{"ZBX_APMTLSCERTFILE": "cert.pem", "ZBX_STARTAPMCOLLECTOR": "1"},
			func(home string) string { return filepath.Join(home, "enc", "cert.pem") },
		},
		{
			"web APM",
			SecretsWebMySQL,
			SecretsService,
			Environment{"ZBX_APM_CAFILE": "ca.pem"},
			func(home string) string { return filepath.Join(home, "ssl", "ssl_ca", "ca.pem") },
		},
		{
			"history provider",
			SecretsServerMySQL,
			SecretsInitHistory,
			Environment{"ZBX_HISTORYPROVIDER_0": provider, "ZBX_HISTORYPROVIDER_0_SSL_CERTFILE": "cert.pem"},
			func(home string) string { return filepath.Join(home, "enc", "cert.pem") },
		},
		{
			"telemetry provider",
			SecretsProxySQLite,
			SecretsInitTelemetry,
			Environment{"ZBX_TELEMETRYPROVIDER_0": provider, "ZBX_TELEMETRYPROVIDER_0_SSL_CERTFILE": "cert.pem"},
			func(home string) string { return filepath.Join(home, "enc", "cert.pem") },
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			external, _, _ := useSecretRoots(t)
			home := t.TempDir()
			writeTestSecret(t, test.legacy(home), "legacy")
			test.env["ZABBIX_USER_HOME_DIR"] = home
			err := ResolveSecrets(test.env, test.component, test.command)
			if err == nil || !strings.Contains(err.Error(), external) {
				t.Fatalf("ResolveSecrets() error=%v, want missing file under %q", err, external)
			}
		})
	}
}

func TestResolveSecretsUsesNewAPMPaths(t *testing.T) {
	external, internal, _ := useSecretRoots(t)
	writeTestSecret(t, filepath.Join(external, "apm", "cert.pem"), "certificate")
	env := Environment{
		"ZABBIX_USER_HOME_DIR":  t.TempDir(),
		"ZBX_APMTLSCERTFILE":    "cert.pem",
		"ZBX_STARTAPMCOLLECTOR": "1",
	}
	if err := ResolveSecrets(env, SecretsProxySQLite, SecretsService); err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(external, "apm", "cert.pem"); env["ZBX_APMTLSCERTFILE"] != want {
		t.Fatalf("ZBX_APMTLSCERTFILE=%q, want %q", env["ZBX_APMTLSCERTFILE"], want)
	}
	if env["ZBX_SSLCALOCATION"] != "" {
		t.Fatalf("ZBX_SSLCALOCATION=%q, want empty", env["ZBX_SSLCALOCATION"])
	}
	if _, err := os.Stat(filepath.Join(internal, "ssl", "ca")); !os.IsNotExist(err) {
		t.Fatalf("global CA output exists without CA files: %v", err)
	}
}

func TestResolveSecretsRemovesInactiveGlobalCA(t *testing.T) {
	_, internal, _ := useSecretRoots(t)
	target := filepath.Join(internal, "ssl", "ca")
	writeTestSecret(t, filepath.Join(target, "stale.pem"), "stale")
	env := Environment{"ZABBIX_USER_HOME_DIR": t.TempDir()}

	if err := ResolveSecrets(env, SecretsServerMySQL, SecretsService); err != nil {
		t.Fatal(err)
	}
	if env["ZBX_SSLCALOCATION"] != "" {
		t.Fatalf("ZBX_SSLCALOCATION=%q, want empty", env["ZBX_SSLCALOCATION"])
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("inactive managed global CA directory was not removed: %v", err)
	}
}

func TestResolveSecretsPreparesConfiguredGlobalCA(t *testing.T) {
	external, internal, _ := useSecretRoots(t)
	writeTestSecret(t, filepath.Join(external, "ssl", "ca", "root.pem"), "certificate")
	env := Environment{"ZABBIX_USER_HOME_DIR": t.TempDir()}

	if err := ResolveSecrets(env, SecretsServerMySQL, SecretsService); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(internal, "ssl", "ca")
	if env["ZBX_SSLCALOCATION"] != target {
		t.Fatalf("ZBX_SSLCALOCATION=%q, want %q", env["ZBX_SSLCALOCATION"], target)
	}
	if data, err := os.ReadFile(filepath.Join(target, "root.pem")); err != nil || string(data) != "certificate" {
		t.Fatalf("global CA=%q, %v", data, err)
	}
}

func TestResolveSecretsDoesNotUseLegacyAPMCADirectory(t *testing.T) {
	_, internal, _ := useSecretRoots(t)
	home := t.TempDir()
	writeTestSecret(t, filepath.Join(home, "ssl", "ssl_ca", "legacy.pem"), "legacy")
	stale := filepath.Join(internal, "apm", "ca", "stale.pem")
	writeTestSecret(t, stale, "stale")
	env := Environment{"ZABBIX_USER_HOME_DIR": home}
	if err := ResolveSecrets(env, SecretsWebMySQL, SecretsService); err != nil {
		t.Fatal(err)
	}
	if env["ZBX_APM_CA_LOCATION"] != "" {
		t.Fatalf("ZBX_APM_CA_LOCATION=%q, want empty", env["ZBX_APM_CA_LOCATION"])
	}
	if _, err := os.Stat(filepath.Dir(stale)); !os.IsNotExist(err) {
		t.Fatalf("inactive managed APM CA directory was not removed: %v", err)
	}
}

func TestResolveSecretsUsesDedicatedAPMCADirectory(t *testing.T) {
	external, internal, _ := useSecretRoots(t)
	writeTestSecret(t, filepath.Join(external, "apm", "ca", "apm.pem"), "apm")
	writeTestSecret(t, filepath.Join(external, "ssl", "ca", "global.pem"), "global")
	env := Environment{"ZABBIX_USER_HOME_DIR": t.TempDir()}
	if err := ResolveSecrets(env, SecretsWebMySQL, SecretsService); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(internal, "apm", "ca")
	if data, err := os.ReadFile(filepath.Join(target, "apm.pem")); err != nil || string(data) != "apm" {
		t.Fatalf("APM CA=%q, %v", data, err)
	}
	if _, err := os.Stat(filepath.Join(target, "global.pem")); !os.IsNotExist(err) {
		t.Fatalf("global CA was copied into the APM CA directory: %v", err)
	}
}

func TestResolveSecretsRejectsSymlinkEscape(t *testing.T) {
	external, _, _ := useSecretRoots(t)
	home := t.TempDir()
	outside := filepath.Join(t.TempDir(), "cert.pem")
	writeTestSecret(t, outside, "outside")
	directory := filepath.Join(external, "tls")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(directory, "cert.pem")); err != nil {
		t.Fatal(err)
	}
	env := Environment{"ZABBIX_USER_HOME_DIR": home, "ZBX_TLSCERTFILE": "cert.pem"}
	err := ResolveSecrets(env, SecretsAgent, SecretsService)
	if err == nil || !strings.Contains(err.Error(), "outside") {
		t.Fatalf("ResolveSecrets() error=%v, want symlink escape error", err)
	}
}

func symlinkOutside(t *testing.T, link string) {
	t.Helper()
	outside := filepath.Join(t.TempDir(), "outside.pem")
	writeTestSecret(t, outside, "outside")
	if err := os.MkdirAll(filepath.Dir(link), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}
}

func TestResolveSecretsRejectsEscape(t *testing.T) {
	const provider = `{"provider":"clickhouse"}`
	for _, test := range []struct {
		name      string
		component SecretComponent
		command   SecretCommand
		env       Environment
		link      func(external, standard, home string) string
		want      string
	}{
		{"file secret", SecretsAgent, SecretsService,
			Environment{"ZBX_TLSCERTFILE": "../cert.pem"}, nil, "escapes"},
		{"value secret", SecretsServerMySQL, SecretsInitDB,
			Environment{"MYSQL_PASSWORD_FILE": "../password"}, nil, "escapes"},
		{"provider file", SecretsServerMySQL, SecretsInitHistory,
			Environment{"ZBX_HISTORYPROVIDER_0": provider, "ZBX_HISTORYPROVIDER_0_SSL_CERTFILE": "a/../../cert.pem"}, nil, "escapes"},
		{"deprecated location symlink", SecretsAgent, SecretsService,
			Environment{"ZBX_TLSCERTFILE": "cert.pem"},
			func(_, _, home string) string { return filepath.Join(home, "enc", "cert.pem") }, "outside"},
		{"docker secrets symlink", SecretsServerMySQL, SecretsInitDB,
			Environment{"ZBX_DBTLSCONNECT": "verify_ca", "ZBX_DBTLSCAFILE": "ca.pem"},
			func(_, standard, _ string) string { return filepath.Join(standard, "ca.pem") }, "outside"},
		{"provider symlink", SecretsServerMySQL, SecretsInitHistory,
			Environment{"ZBX_HISTORYPROVIDER_0": provider, "ZBX_HISTORYPROVIDER_0_SSL_CERTFILE": "cert.pem"},
			func(external, _, _ string) string {
				return filepath.Join(external, "providers", "history-0", "cert.pem")
			}, "outside"},
		{"CA directory symlink", SecretsServerMySQL, SecretsService,
			Environment{},
			func(external, _, _ string) string { return filepath.Join(external, "ssl", "ca", "root.pem") }, "outside"},
	} {
		t.Run(test.name, func(t *testing.T) {
			external, _, standard := useSecretRoots(t)
			home := t.TempDir()
			if test.link != nil {
				symlinkOutside(t, test.link(external, standard, home))
			}
			test.env["ZABBIX_USER_HOME_DIR"] = home
			err := ResolveSecrets(test.env, test.component, test.command)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("ResolveSecrets() error=%v, want %q", err, test.want)
			}
		})
	}
}

// Kubernetes Secret volumes expose keys as symlinks into ..data.
func TestResolveSecretsAcceptsProjectedVolumeLayout(t *testing.T) {
	external, internal, _ := useSecretRoots(t)
	for _, directory := range []string{filepath.Join(external, "tls"), filepath.Join(external, "ssl", "ca")} {
		writeTestSecret(t, filepath.Join(directory, "..2026_11_06", "file.pem"), "data")
		if err := os.Symlink("..2026_11_06", filepath.Join(directory, "..data")); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(filepath.Join("..data", "file.pem"), filepath.Join(directory, "file.pem")); err != nil {
			t.Fatal(err)
		}
	}

	env := Environment{"ZABBIX_USER_HOME_DIR": t.TempDir(), "ZBX_TLSCERTFILE": "file.pem"}
	if err := ResolveSecrets(env, SecretsAgent, SecretsService); err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(external, "tls", "file.pem"); env["ZBX_TLSCERTFILE"] != want {
		t.Fatalf("ZBX_TLSCERTFILE=%q, want %q", env["ZBX_TLSCERTFILE"], want)
	}

	env = Environment{"ZABBIX_USER_HOME_DIR": t.TempDir()}
	if err := ResolveSecrets(env, SecretsServerMySQL, SecretsService); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(filepath.Join(internal, "ssl", "ca", "file.pem")); err != nil || string(data) != "data" {
		t.Fatalf("managed CA file=%q, %v", data, err)
	}
}

func TestResolveSecretsNormalizesAliases(t *testing.T) {
	_, _, _ = useSecretRoots(t)
	cert := filepath.Join(t.TempDir(), "vault.pem")
	writeTestSecret(t, cert, "cert")
	env := Environment{"ZBX_VAULT": "HashiCorp", "ZBX_VAULTCERTFILE": cert}
	if err := ResolveSecrets(env, SecretsServerMySQL, SecretsInitDB); err != nil {
		t.Fatal(err)
	}
	if env["ZBX_VAULTTLSCERTFILE"] != cert {
		t.Fatalf("canonical alias=%q, want %q", env["ZBX_VAULTTLSCERTFILE"], cert)
	}
	if _, exists := env["ZBX_VAULTCERTFILE"]; exists {
		t.Fatal("legacy alias was not removed")
	}
}

func TestResolveSecretsUsesVaultTLSFileNamesForServer(t *testing.T) {
	external, _, _ := useSecretRoots(t)
	cert := filepath.Join(external, "ssl", "certs", "client.pem")
	key := filepath.Join(external, "ssl", "keys", "client.key")
	writeTestSecret(t, cert, "certificate")
	writeTestSecret(t, key, "private key")

	env := Environment{
		"ZABBIX_USER_HOME_DIR": t.TempDir(),
		"ZBX_VAULT":            "CyberArk",
		"ZBX_VAULTTLSCERTFILE": "client.pem",
		"ZBX_VAULTTLSKEYFILE":  "client.key",
	}
	if err := ResolveSecrets(env, SecretsServerMySQL, SecretsService); err != nil {
		t.Fatal(err)
	}

	if env["ZBX_VAULTTLSCERTFILE"] != cert || env["ZBX_VAULTTLSKEYFILE"] != key {
		t.Fatalf("resolved Vault TLS paths=%q/%q, want %q/%q", env["ZBX_VAULTTLSCERTFILE"], env["ZBX_VAULTTLSKEYFILE"], cert, key)
	}
	if err := SetVaultTLSFileNames(env); err != nil {
		t.Fatal(err)
	}
	if env["ZBX_VAULTTLSCERTFILE"] != "client.pem" || env["ZBX_VAULTTLSKEYFILE"] != "client.key" {
		t.Fatalf("Vault TLS files=%q/%q, want file names", env["ZBX_VAULTTLSCERTFILE"], env["ZBX_VAULTTLSKEYFILE"])
	}
}

func TestResolveSecretsKeepsVaultTLSPathForWeb(t *testing.T) {
	external, _, _ := useSecretRoots(t)
	cert := filepath.Join(external, "vault", "client.pem")
	writeTestSecret(t, cert, "certificate")
	env := Environment{"ZBX_VAULT": "HashiCorp", "ZBX_VAULTTLSCERTFILE": "client.pem"}

	if err := ResolveSecrets(env, SecretsWebMySQL, SecretsService); err != nil {
		t.Fatal(err)
	}
	if env["ZBX_VAULTTLSCERTFILE"] != cert {
		t.Fatalf("ZBX_VAULTTLSCERTFILE=%q, want %q", env["ZBX_VAULTTLSCERTFILE"], cert)
	}
}

func TestResolveSecretsCopiesVaultTLSFilesOutsideSharedLocations(t *testing.T) {
	external, internal, _ := useSecretRoots(t)
	writeTestSecret(t, filepath.Join(external, "ssl", "certs", "web.pem"), "web certificate")
	writeTestSecret(t, filepath.Join(external, "ssl", "keys", "web.key"), "web private key")
	cert := filepath.Join(t.TempDir(), "vault.pem")
	writeTestSecret(t, cert, "vault certificate")

	env := Environment{
		"ZABBIX_USER_HOME_DIR": "ignored",
		"ZBX_VAULT":            "CyberArk",
		"ZBX_VAULTTLSCERTFILE": cert,
		"ZBX_VAULTTLSKEY":      "vault private key",
	}
	if err := ResolveSecrets(env, SecretsProxyMySQL, SecretsService); err != nil {
		t.Fatal(err)
	}

	certDir := filepath.Join(internal, "ssl", "certs")
	keyDir := filepath.Join(internal, "ssl", "keys")
	if env["ZBX_SSLCERTLOCATION"] != certDir || env["ZBX_SSLKEYLOCATION"] != keyDir {
		t.Fatalf("SSL locations=%q/%q, want %q/%q", env["ZBX_SSLCERTLOCATION"], env["ZBX_SSLKEYLOCATION"], certDir, keyDir)
	}
	for path, want := range map[string]string{
		filepath.Join(certDir, "web.pem"):              "web certificate",
		filepath.Join(certDir, "ZBX_VAULTTLSCERTFILE"): "vault certificate",
		filepath.Join(keyDir, "web.key"):               "web private key",
		filepath.Join(keyDir, "ZBX_VAULTTLSKEYFILE"):   "vault private key",
	} {
		data, err := os.ReadFile(path)
		if err != nil || string(data) != want {
			t.Fatalf("%s=%q, %v; want %q", path, data, err, want)
		}
	}
	if env["ZBX_VAULTTLSCERTFILE"] != filepath.Join(certDir, "ZBX_VAULTTLSCERTFILE") ||
		env["ZBX_VAULTTLSKEYFILE"] != filepath.Join(keyDir, "ZBX_VAULTTLSKEYFILE") {
		t.Fatalf("resolved Vault TLS paths=%q/%q", env["ZBX_VAULTTLSCERTFILE"], env["ZBX_VAULTTLSKEYFILE"])
	}
	if err := SetVaultTLSFileNames(env); err != nil {
		t.Fatal(err)
	}
	if env["ZBX_VAULTTLSCERTFILE"] != "ZBX_VAULTTLSCERTFILE" || env["ZBX_VAULTTLSKEYFILE"] != "ZBX_VAULTTLSKEYFILE" {
		t.Fatalf("Vault TLS files=%q/%q, want managed file names", env["ZBX_VAULTTLSCERTFILE"], env["ZBX_VAULTTLSKEYFILE"])
	}
}

func TestWithResolvedSecretsSkipsCustomCommands(t *testing.T) {
	for _, test := range []struct {
		name      string
		component SecretComponent
		binary    string
		args      []string
	}{
		{"binary image", SecretsMCPServer, "/usr/sbin/zabbix_mcp_server", []string{"sh"}},
		{"web image", SecretsWebMySQL, "", []string{"-f"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			env := Environment{"ZBX_TLSCERT": "certificate"}
			called := false
			entry := WithResolvedSecrets(test.component, test.binary, func(got Environment, args []string) error {
				called = true
				if got["ZBX_TLSCERT"] != "certificate" {
					t.Fatal("custom command environment was modified")
				}
				return nil
			})
			if err := entry(env, test.args); err != nil {
				t.Fatal(err)
			}
			if !called {
				t.Fatal("custom command was not forwarded")
			}
		})
	}
}

func TestInitSecretWrappersPreserveUnsupportedHandlers(t *testing.T) {
	if WithInitDBSecrets(SecretsServerMySQL, nil) != nil {
		t.Fatal("nil database handler became supported")
	}
	if WithHistoryInitSecrets(SecretsServerMySQL, nil) != nil {
		t.Fatal("nil history handler became supported")
	}
	if WithTelemetryInitSecrets(SecretsServerMySQL, nil) != nil {
		t.Fatal("nil telemetry handler became supported")
	}
}

func TestInitSecretWrappersResolveSelectedPlan(t *testing.T) {
	_, _, _ = useSecretRoots(t)
	password := filepath.Join(t.TempDir(), "password")
	writeTestSecret(t, password, "secret")

	dbEnv := Environment{"MYSQL_PASSWORD_FILE": password}
	dbCalled := false
	db := WithInitDBSecrets(SecretsServerMySQL, func(env Environment) error {
		dbCalled = true
		if env["MYSQL_PASSWORD"] != "secret" {
			t.Fatalf("database password=%q", env["MYSQL_PASSWORD"])
		}
		return nil
	})
	if err := db(dbEnv); err != nil {
		t.Fatal(err)
	}
	if !dbCalled {
		t.Fatal("database handler was not called")
	}

	historyEnv := Environment{
		"ZBX_HISTORYPROVIDER_0":               `{"provider":"clickhouse"}`,
		"ZBX_HISTORYPROVIDER_0_PASSWORD_FILE": password,
		"ZBX_SSLCALOCATION":                   t.TempDir(),
	}
	history := WithHistoryInitSecrets(SecretsServerMySQL, func(env Environment) (int, error) {
		if env["ZBX_HISTORYPROVIDER_0_PASSWORD"] != "secret" {
			t.Fatalf("history password=%q", env["ZBX_HISTORYPROVIDER_0_PASSWORD"])
		}
		return 1, nil
	})
	if count, err := history(historyEnv); err != nil || count != 1 {
		t.Fatalf("history handler count=%d, error=%v", count, err)
	}
}

func TestWebHistoryProviderValueSecretUsesProviderCategory(t *testing.T) {
	external, _, _ := useSecretRoots(t)
	writeTestSecret(t, filepath.Join(external, "providers", "history-0", "username"), "history-user")
	env := Environment{
		"ZBX_HISTORYPROVIDER_0":               `{"provider":"clickhouse"}`,
		"ZBX_HISTORYPROVIDER_0_USERNAME_FILE": "username",
		"ZBX_APM_CA_LOCATION":                 filepath.Join(t.TempDir(), "custom-ca"),
	}
	if err := ResolveSecrets(env, SecretsWebMySQL, SecretsService); err != nil {
		t.Fatal(err)
	}
	if env["ZBX_HISTORYPROVIDER_0_USERNAME"] != "history-user" {
		t.Fatalf("history username=%q", env["ZBX_HISTORYPROVIDER_0_USERNAME"])
	}
}
