//go:build !windows

package bootstrap

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestResolveDBTLS(t *testing.T) {
	env := Environment{"ZBX_DBTLSCONNECT": "verify_ca", "ZBX_DBTLSCAFILE": "ca.pem"}
	config, err := ResolveServiceDBTLS(env)
	if err != nil {
		t.Fatal(err)
	}
	if want := "/run/secrets/ca.pem"; config.CAFile != want || env["ZBX_DBTLSCAFILE"] != want {
		t.Fatalf("resolved DB TLS file = %q, environment = %q", config.CAFile, env["ZBX_DBTLSCAFILE"])
	}

	env = Environment{"ZBX_DBTLSCONNECT": "verify_ca", "ZBX_DBTLSCAFILE": "../ca.pem"}
	if _, err := ResolveServiceDBTLS(env); err == nil {
		t.Fatal("escaping DB TLS path was accepted")
	}

	env = Environment{"ZBX_DBTLSCAFILE": "missing.pem"}
	config, err = ResolveServiceDBTLS(env)
	if err != nil {
		t.Fatalf("disabled DB TLS validated an unused file: %v", err)
	}
	if config.ConnectMode != "" || config.CAFile != "missing.pem" || env[envDBTLSCAFile] != "missing.pem" {
		t.Fatalf("disabled DB TLS changed file settings: config=%+v, env=%v", config, env)
	}
}

func TestDBTLSByComponent(t *testing.T) {
	env := Environment{
		"ZBX_DBTLSCONNECT":   "verify_ca",
		"ZBX_DBTLSCAFILE":    "/shared/ca.pem",
		"ZBX_DBTLSCERTFILE":  "/shared/cert.pem",
		"ZBX_DBTLSKEYFILE":   "/shared/key.pem",
		"ZBX_DB_ENCRYPTION":  "true",
		"ZBX_DB_VERIFY_HOST": "true",
	}

	service, err := ResolveServiceDBTLS(env)
	if err != nil {
		t.Fatal(err)
	}
	wantService := DBTLSConfig{
		ConnectMode: "verify_ca",
		CAFile:      "/shared/ca.pem",
		CertFile:    "/shared/cert.pem",
		KeyFile:     "/shared/key.pem",
	}
	if !reflect.DeepEqual(service, wantService) {
		t.Fatalf("service TLS settings = %#v, want %#v", service, wantService)
	}

	frontend, err := ResolveFrontendDBTLS(env)
	if err != nil {
		t.Fatal(err)
	}
	wantFrontend := DBTLSConfig{
		ConnectMode: "verify_full",
		CAFile:      "/shared/ca.pem",
		CertFile:    "/shared/cert.pem",
		KeyFile:     "/shared/key.pem",
	}
	if !reflect.DeepEqual(frontend, wantFrontend) {
		t.Fatalf("frontend TLS settings = %#v, want %#v", frontend, wantFrontend)
	}
}

func TestFrontendDBTLS(t *testing.T) {
	for _, test := range []struct {
		name string
		env  Environment
		mode string
	}{
		{name: "disabled", env: Environment{"ZBX_DB_ENCRYPTION": "false"}},
		{name: "encryption only", env: Environment{"ZBX_DB_ENCRYPTION": "TRUE"}, mode: "required"},
		{
			name: "verify CA",
			env:  Environment{"ZBX_DB_ENCRYPTION": "true", "ZBX_DBTLSCAFILE": "/ca.pem"},
			mode: "verify_ca",
		},
		{
			name: "verify identity",
			env: Environment{
				"ZBX_DB_ENCRYPTION": "true", "ZBX_DB_VERIFY_HOST": "true", "ZBX_DBTLSCAFILE": "/ca.pem",
			},
			mode: "verify_full",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			settings, err := ResolveFrontendDBTLS(test.env)
			if err != nil {
				t.Fatal(err)
			}
			if settings.ConnectMode != test.mode {
				t.Fatalf("connection mode = %q, want %q", settings.ConnectMode, test.mode)
			}
		})
	}
}

func TestResolveDBPort(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		want      string
		wantError bool
	}{
		{name: "default", want: "5432"},
		{name: "configured", value: "15432", want: "15432"},
		{name: "maximum", value: "65535", want: "65535"},
		{name: "not numeric", value: "postgresql", wantError: true},
		{name: "out of range", value: "65536", wantError: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env := Environment{}
			if test.value != "" {
				env["DB_SERVER_PORT"] = test.value
			}

			got, err := ResolveDBPort(env, "5432")
			if (err != nil) != test.wantError {
				t.Fatalf("ResolveDBPort() error = %v, wantError = %t", err, test.wantError)
			}
			if test.wantError {
				if !strings.Contains(err.Error(), "DB_SERVER_PORT") {
					t.Fatalf("ResolveDBPort() error = %v, want DB_SERVER_PORT context", err)
				}
				return
			}
			if got != test.want || env["DB_SERVER_PORT"] != test.want {
				t.Fatalf("ResolveDBPort() = %q, env value = %q, want %q", got, env["DB_SERVER_PORT"], test.want)
			}
		})
	}
}

func TestInlineDBTLS(t *testing.T) {
	runtimeDir := useDBTLSRuntimeDir(t)
	env := Environment{
		"ZBX_DBTLSCONNECT":  "verify_full",
		"ZBX_DBTLSCA":       "inline CA",
		"ZBX_DBTLSCERTFILE": "client.pem",
		"ZBX_DBTLSKEY":      "inline key",
	}
	config, err := ResolveServiceDBTLS(env)
	if err != nil {
		t.Fatal(err)
	}
	want := DBTLSConfig{
		ConnectMode: "verify_full",
		CAFile:      filepath.Join(runtimeDir, "ca.pem"),
		CertFile:    "/run/secrets/client.pem",
		KeyFile:     filepath.Join(runtimeDir, "key.pem"),
	}
	if !reflect.DeepEqual(config, want) {
		t.Fatalf("DB TLS = %#v, want %#v", config, want)
	}
	for path, content := range map[string]string{want.CAFile: "inline CA", want.KeyFile: "inline key"} {
		data, err := os.ReadFile(path)
		if err != nil || string(data) != content {
			t.Errorf("%s = %q, %v; want %q", path, data, err, content)
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("stat %s: %v", path, err)
		} else if info.Mode().Perm() != 0o600 {
			t.Errorf("%s mode = %o, want 600", path, info.Mode().Perm())
		}
	}
	for _, name := range []string{"ZBX_DBTLSCA", "ZBX_DBTLSKEY"} {
		if _, found := env[name]; found {
			t.Errorf("inline %s was left in the environment", name)
		}
	}
}

func TestInlineDBTLSFrontendModeAndConflicts(t *testing.T) {
	runtimeDir := useDBTLSRuntimeDir(t)
	env := Environment{
		"ZBX_DB_ENCRYPTION": "true", "ZBX_DB_VERIFY_HOST": "true", "ZBX_DBTLSCA": "inline CA",
	}
	config, err := ResolveFrontendDBTLS(env)
	if err != nil {
		t.Fatal(err)
	}
	if config.ConnectMode != "verify_full" || config.CAFile != filepath.Join(runtimeDir, "ca.pem") {
		t.Fatalf("frontend DB TLS with inline CA = %#v", config)
	}

	env = Environment{"ZBX_DBTLSCONNECT": "required", "ZBX_DBTLSCERT": "inline", "ZBX_DBTLSCERTFILE": "cert.pem"}
	if _, err := ResolveServiceDBTLS(env); err == nil || !strings.Contains(err.Error(), "exclusive") {
		t.Fatalf("inline and file DB TLS certificate: error = %v", err)
	}

}

func TestInlineDBTLSConflictDoesNotWriteFiles(t *testing.T) {
	runtimeDir := useDBTLSRuntimeDir(t)
	env := Environment{
		"ZBX_DBTLSCONNECT":  "required",
		"ZBX_DBTLSCA":       "inline CA",
		"ZBX_DBTLSCERT":     "inline certificate",
		"ZBX_DBTLSCERTFILE": "cert.pem",
	}
	if _, err := ResolveServiceDBTLS(env); err == nil || !strings.Contains(err.Error(), "exclusive") {
		t.Fatalf("inline and file DB TLS certificate: error = %v", err)
	}
	if _, err := os.Stat(runtimeDir); !os.IsNotExist(err) {
		t.Fatalf("DB TLS runtime directory was created before validation: %v", err)
	}
}

func TestInlineDBTLSRemovedWhenDisabled(t *testing.T) {
	for _, test := range []struct {
		name    string
		resolve func(Environment) (DBTLSConfig, error)
		env     Environment
	}{
		{"service", ResolveServiceDBTLS, Environment{}},
		{"frontend", ResolveFrontendDBTLS, Environment{"ZBX_DB_ENCRYPTION": "false"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			runtimeDir := useDBTLSRuntimeDir(t)
			env := test.env
			env["ZBX_DBTLSCA"] = "inline CA"
			env["ZBX_DBTLSCERTFILE"] = "client.pem"
			config, err := test.resolve(env)
			if err != nil {
				t.Fatal(err)
			}
			if config.ConnectMode != "" {
				t.Errorf("connection mode = %q, want disabled", config.ConnectMode)
			}
			if _, found := env["ZBX_DBTLSCA"]; found {
				t.Error("inline ZBX_DBTLSCA was left in the environment")
			}
			if want := filepath.Join(runtimeDir, "ca.pem"); env["ZBX_DBTLSCAFILE"] != want {
				t.Errorf("ZBX_DBTLSCAFILE = %q, want %q", env["ZBX_DBTLSCAFILE"], want)
			}
			wantCertFile := "client.pem"
			if test.name == "frontend" {
				wantCertFile = filepath.Join(SecretsDir, "client.pem")
			}
			if env["ZBX_DBTLSCERTFILE"] != wantCertFile {
				t.Errorf("disabled DB TLS changed ZBX_DBTLSCERTFILE to %q, want %q", env["ZBX_DBTLSCERTFILE"], wantCertFile)
			}
		})
	}
}

func useDBTLSRuntimeDir(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "database")
	previous := dbTLSRuntimeDir
	dbTLSRuntimeDir = dir
	t.Cleanup(func() { dbTLSRuntimeDir = previous })
	return dir
}
