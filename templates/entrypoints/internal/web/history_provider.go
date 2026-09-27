package web

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/provider"
)

// escapedApostrophe is the JSON escape of an apostrophe. The generic PHP
// env_json helper also accepts single-quoted JSON, so an apostrophe inside
// a value must stay escaped until PHP decodes it.
const escapedApostrophe = "\\u0027"

// configureWebHistoryProviders turns the unindexed and indexed image variables
// into one JSON array for the PHP configuration. Credentials supplied through
// separate environment variables take precedence over JSON fields.
func configureWebHistoryProviders(env bootstrap.Environment) error {
	providers, err := provider.WebProviders(env, provider.History)
	if err != nil {
		return err
	}
	return setWebProviders(env, provider.History.Plural, providers)
}

// setWebProviders stores the providers as the JSON array read by the PHP
// configuration.
func setWebProviders(env bootstrap.Environment, name string, providers []map[string]any) error {
	encoded, err := json.Marshal(providers)
	if err != nil {
		return fmt.Errorf("encode %s: %w", name, err)
	}
	env[name] = strings.ReplaceAll(string(encoded), "'", escapedApostrophe)
	return nil
}
