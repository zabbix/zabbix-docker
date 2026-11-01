package storageprovider

import "strings"

// telemetryPrefix is the variable prefix of one telemetry data source.
const telemetryPrefix = "ZBX_TELEMETRYPROVIDER"

var clickHouseTelemetry = providerDefinition{
	files: tlsFiles,
	web: []string{
		"provider", "url", "db", "vault_path", "ssl_cert_file", "ssl_key_file",
		"ssl_key_password", "ssl_verify_peer", "ssl_verify_host",
		"username", "password",
	},
	// The frontend distinguishes empty strings from missing options (null).
	webDefaults: map[string]any{
		"url": "", "db": "", "username": "", "password": "", "vault_path": "",
		"ssl_cert_file": "", "ssl_key_file": "", "ssl_key_password": "",
		"ssl_ca_file": "", "ssl_ca_location": "",
		"ssl_verify_peer": false, "ssl_verify_host": false,
	},
	secrets: true,
}

// Telemetry supports one indexed provider; the plural variable is output-only.
var Telemetry = &Parameter{
	Prefix:        telemetryPrefix,
	Plural:        "ZBX_TELEMETRYPROVIDERS",
	maxProviders:  1,
	providerError: "provider must be clickhouse",
	providers:     map[string]providerDefinition{"clickhouse": clickHouseTelemetry},
	defaults:      telemetryDefaults,
}

// The frontend needs explicit ssl_verify_peer to select the CA directory.
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
