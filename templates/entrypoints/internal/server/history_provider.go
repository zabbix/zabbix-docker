package server

import (
	"fmt"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/provider"
)

// configureJSONHistoryProviders converts JSON values to the native server
// format. Existing native values remain unchanged.
func configureJSONHistoryProviders(env bootstrap.Environment) error {
	if _, exists := env[provider.History.Plural]; exists {
		return fmt.Errorf("%s is not supported by the server; use %s",
			provider.History.Plural, bootstrap.IndexedName(provider.HistoryPrefix, 0))
	}
	return provider.ConfigureNative(env, provider.History)
}
