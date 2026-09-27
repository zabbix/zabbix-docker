package provider

import (
	"strings"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

// TelemetryPrefix is the variable prefix of one telemetry data source.
const TelemetryPrefix = "ZBX_TELEMETRYPROVIDER"

// clickHouseTelemetry describes the options handled by the image and frontend.
var clickHouseTelemetry = schema{
	files: tlsFiles,
	web: []string{
		"provider", "url", "db", "vault_path", "ssl_cert_file", "ssl_key_file",
		"ssl_key_password", "ssl_verify_peer", "ssl_verify_host",
		"username", "password",
	},
	// The frontend reads every option of a telemetry data source without a
	// default of its own, so an option it does not receive is read as null: a
	// missing vault_path selects Vault, and missing credentials send an empty
	// user name. These are the defaults documented for the frontend.
	webDefaults: map[string]any{
		"url": "", "db": "", "username": "", "password": "", "vault_path": "",
		"ssl_cert_file": "", "ssl_key_file": "", "ssl_key_password": "",
		"ssl_ca_file": "", "ssl_ca_location": "",
		"ssl_verify_peer": false, "ssl_verify_host": false,
	},
	secrets: true,
}

// Telemetry describes the server and proxy TelemetryProvider parameter and the
// frontend telemetry providers. Only one data source is supported, and the
// frontend variable is generated from it instead of being read.
var Telemetry = &parameter{
	Prefix:        TelemetryPrefix,
	Plural:        "ZBX_TELEMETRYPROVIDERS",
	maxProviders:  1,
	providerError: "provider must be clickhouse",
	reserved:      managedLocations,
	providers:     map[string]schema{"clickhouse": clickHouseTelemetry},
	defaults:      telemetryDefaults,
}

// ConfigureNativeTelemetry converts the telemetry provider of the server or
// proxy to the native configuration format.
func ConfigureNativeTelemetry(env bootstrap.Environment) error {
	return ConfigureNative(env, Telemetry)
}

// telemetryDefaults verifies the certificate of an encrypted connection unless
// the configuration says otherwise. The frontend derives the CA directory from
// ssl_verify_peer, so the default has to be explicit.
func telemetryDefaults(fields map[string]any) {
	if rawURL, _ := fields["url"].(string); !strings.HasPrefix(strings.ToLower(rawURL), "https://") {
		return
	}
	for _, key := range []string{"ssl_verify_peer", "ssl_verify_host"} {
		if _, exists := fields[key]; !exists {
			fields[key] = true
		}
	}
}
