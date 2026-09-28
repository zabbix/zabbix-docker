package bootstrap

import (
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
	if _, err := ResolveServiceDBTLS(env); err != nil {
		t.Fatalf("disabled DB TLS validated an unused file: %v", err)
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
