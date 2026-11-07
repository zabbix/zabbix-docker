package web

import (
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/storageprovider"
)

// configureWebTelemetryProviders generates the frontend telemetry
// configuration from the image variables. Relative deprecated JSON paths use
// the canonical TLS secret directories without a legacy fallback.
func configureWebTelemetryProviders(env bootstrap.Environment) error {
	providers, err := storageprovider.WebProviders(env, storageprovider.Telemetry)
	if err != nil {
		return err
	}
	for _, fields := range providers {
		for field, category := range map[string]string{"ssl_cert_file": "ssl/certs", "ssl_key_file": "ssl/keys"} {
			if value, _ := fields[field].(string); value != "" {
				path, err := bootstrap.ResolveExternalSecretFile(field, value, category)
				if err != nil {
					return err
				}
				fields[field] = path
			}
		}
	}
	return setWebProviders(env, storageprovider.Telemetry.Plural, providers)
}
