package bootstrap

import (
	"reflect"
	"strings"
	"testing"
)

func TestDBTLSByComponent(t *testing.T) {
	env := Environment{
		"ZBX_DBTLSCONNECT":   "verify_ca",
		"ZBX_DBTLSCAFILE":    "/service/ca.pem",
		"ZBX_DBTLSCERTFILE":  "/service/cert.pem",
		"ZBX_DBTLSKEYFILE":   "/service/key.pem",
		"ZBX_DB_ENCRYPTION":  "true",
		"ZBX_DB_VERIFY_HOST": "true",
		"ZBX_DB_CA_FILE":     "/frontend/ca.pem",
		"ZBX_DB_CERT_FILE":   "/frontend/cert.pem",
		"ZBX_DB_KEY_FILE":    "/frontend/key.pem",
	}

	service := ServiceDBTLS(env)
	wantService := DBTLSConfig{
		ConnectMode: "verify_ca",
		CAFile:      "/service/ca.pem",
		CertFile:    "/service/cert.pem",
		KeyFile:     "/service/key.pem",
	}
	if !reflect.DeepEqual(service, wantService) {
		t.Fatalf("service TLS settings = %#v, want %#v", service, wantService)
	}

	frontend := FrontendDBTLS(env)
	wantFrontend := DBTLSConfig{
		ConnectMode: "verify_full",
		CAFile:      "/frontend/ca.pem",
		CertFile:    "/frontend/cert.pem",
		KeyFile:     "/frontend/key.pem",
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
			env:  Environment{"ZBX_DB_ENCRYPTION": "true", "ZBX_DB_CA_FILE": "/ca.pem"},
			mode: "verify_ca",
		},
		{
			name: "verify identity",
			env: Environment{
				"ZBX_DB_ENCRYPTION": "true", "ZBX_DB_VERIFY_HOST": "true", "ZBX_DB_CA_FILE": "/ca.pem",
			},
			mode: "verify_full",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			settings := FrontendDBTLS(test.env)
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
