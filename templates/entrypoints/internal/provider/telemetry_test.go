package provider

import (
	"strings"
	"testing"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

func TestConfigureNativeTelemetry(t *testing.T) {
	env := bootstrap.Environment{
		"ZBX_TELEMETRYPROVIDER_0":                  `{"provider":"clickhouse","url":"https://clickhouse:8443","db":"zabbix","ssl_cert_file":"client.crt","ssl_key_file":"client.key"}`,
		"ZBX_TELEMETRYPROVIDER_0_USERNAME":         "zabbix",
		"ZBX_TELEMETRYPROVIDER_0_PASSWORD":         "secret",
		"ZBX_TELEMETRYPROVIDER_0_SSL_KEY_PASSWORD": "key-secret",
	}
	if err := ConfigureNativeTelemetry(env); err != nil {
		t.Fatal(err)
	}
	value := env["ZBX_TELEMETRYPROVIDER_0"]
	for _, expected := range []string{
		`ssl_cert_file="client.crt"`,
		`ssl_key_file="client.key"`,
		`ssl_key_password="key-secret"`,
		"ssl_verify_peer=1",
		"ssl_verify_host=1",
		`username="zabbix"`,
		`password="secret"`,
	} {
		if !strings.Contains(value, expected) {
			t.Errorf("provider %q does not contain %q", value, expected)
		}
	}
	for _, forbidden := range []string{"ssl_ca_location", "ssl_cert_location", "ssl_key_location"} {
		if strings.Contains(value, forbidden) {
			t.Errorf("provider contains managed option %s: %q", forbidden, value)
		}
	}
	for _, secret := range ClickHouseSecretFields {
		if _, exists := env["ZBX_TELEMETRYPROVIDER_0"+secret.Suffix]; exists {
			t.Errorf("%s remains in the environment", secret.Suffix)
		}
	}
}

func TestTelemetryProviderValidation(t *testing.T) {
	tests := []struct {
		name    string
		env     bootstrap.Environment
		wantErr string
	}{
		{
			name: "multiple providers",
			env: bootstrap.Environment{
				"ZBX_TELEMETRYPROVIDER_0": `{"provider":"clickhouse","url":"http://one:8123","db":"zabbix"}`,
				"ZBX_TELEMETRYPROVIDER_1": `{"provider":"clickhouse","url":"http://two:8123","db":"zabbix"}`,
			},
			wantErr: "ZBX_TELEMETRYPROVIDER_1 is not supported",
		},
		{
			name: "managed location",
			env: bootstrap.Environment{
				"ZBX_TELEMETRYPROVIDER_0": `{"provider":"clickhouse","url":"https://clickhouse:8443","db":"zabbix","ssl_ca_location":"/ca"}`,
			},
			wantErr: "managed by the image",
		},
		{
			name: "certificate outside default directory",
			env: bootstrap.Environment{
				"ZBX_TELEMETRYPROVIDER_0": `{"provider":"clickhouse","url":"https://clickhouse:8443","db":"zabbix","ssl_cert_file":"../client.crt"}`,
			},
			wantErr: "escapes the default directory",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ConfigureNativeTelemetry(test.env)
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("ConfigureNativeTelemetry() error = %v, want %q", err, test.wantErr)
			}
		})
	}
}

func TestTelemetryDefaultsKeepPlainConnections(t *testing.T) {
	env := bootstrap.Environment{
		"ZBX_TELEMETRYPROVIDER_0": `{"provider":"clickhouse","url":"http://clickhouse:8123","db":"zabbix"}`,
	}
	if err := ConfigureNativeTelemetry(env); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(env["ZBX_TELEMETRYPROVIDER_0"], "ssl_verify") {
		t.Fatalf("plain connection was given verification defaults: %q", env["ZBX_TELEMETRYPROVIDER_0"])
	}
}
