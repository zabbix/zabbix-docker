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
		for field, directory := range map[string]string{"ssl_cert_file": "certs", "ssl_key_file": "keys"} {
			if value, _ := fields[field].(string); value != "" {
				path, err := bootstrap.ResolveFile(value, filepath.Join(homeDir, "ssl", directory))
				if err != nil {
					return err
				}
				fields[field] = path
			}
		}
		if verifyPeer, _ := fields["ssl_verify_peer"].(bool); verifyPeer {
			fields["ssl_ca_location"] = filepath.Join(homeDir, "ssl", "ssl_ca")
		}
	}
	return setWebProviders(env, provider.Telemetry.Plural, providers)
}
