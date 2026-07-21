package bootstrap

import (
	"reflect"
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

func TestApplyDBCredentials(t *testing.T) {
	tests := []struct {
		name     string
		env      Environment
		explicit bool
	}{
		{name: "environment", env: Environment{}, explicit: true},
		{
			name: "vault without database path",
			env: Environment{
				"ZBX_VAULT": "HashiCorp", "ZBX_VAULTURL": "https://vault.example.test",
			},
			explicit: true,
		},
		{
			name: "vault database credentials",
			env: Environment{
				"ZBX_VAULT": "HashiCorp", "ZBX_VAULTURL": "https://vault.example.test",
				"ZBX_VAULTDBPATH": "database/zabbix",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.env["ZBX_DB_USER"] = "old-user"
			test.env["ZBX_DB_PASSWORD"] = "old-password"

			ApplyDBCredentials(test.env, "zabbix", "secret")

			if test.explicit {
				if test.env["ZBX_DB_USER"] != "zabbix" || test.env["ZBX_DB_PASSWORD"] != "secret" {
					t.Fatalf("database credentials were not applied: %#v", test.env)
				}
				return
			}

			if _, found := test.env["ZBX_DB_USER"]; found {
				t.Fatalf("database user was retained: %#v", test.env)
			}
			if _, found := test.env["ZBX_DB_PASSWORD"]; found {
				t.Fatalf("database password was retained: %#v", test.env)
			}
		})
	}
}
