package web

import (
	"path/filepath"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/provider"
)

// configureWebTelemetryProviders generates the frontend telemetry
// configuration from the image variables. TLS material is referenced by the
// directories the image owns, so the frontend reads the same files as the
// server.
func configureWebTelemetryProviders(env bootstrap.Environment, homeDir string) error {
	providers, err := provider.WebProviders(env, provider.Telemetry)
	if err != nil {
		return err
	}
	for _, fields := range providers {
		if value, _ := fields["ssl_cert_file"].(string); value != "" {
			fields["ssl_cert_file"] = filepath.Join(homeDir, "ssl", "certs", value)
		}
		if value, _ := fields["ssl_key_file"].(string); value != "" {
			fields["ssl_key_file"] = filepath.Join(homeDir, "ssl", "keys", value)
		}
		if verifyPeer, _ := fields["ssl_verify_peer"].(bool); verifyPeer {
			fields["ssl_ca_location"] = filepath.Join(homeDir, "ssl", "ssl_ca")
		}
	}
	return setWebProviders(env, provider.Telemetry.Plural, providers)
}
