package storageprovider

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

var providerSecretDefinitions = bootstrap.ProviderSecretDefinitions()

var tlsFiles = providerTLSFiles()

func providerTLSFiles() []string {
	var fields []string
	for _, secret := range providerSecretDefinitions {
		if secret.Kind == bootstrap.ProviderFileSecret && !secret.ManagedLocation {
			fields = append(fields, secret.Field)
		}
	}
	return fields
}

func looksLikeJSONObject(raw string) bool {
	return strings.HasPrefix(strings.TrimSpace(raw), "{")
}

func decodeObject(raw string) (map[string]any, error) {
	value, err := decodeJSON(raw)
	if err != nil {
		return nil, err
	}
	fields, ok := value.(map[string]any)
	if !ok {
		return nil, errors.New("invalid provider JSON")
	}
	return fields, nil
}

func decodeList(raw string) ([]map[string]any, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	value, err := decodeJSON(raw)
	if err != nil {
		return nil, err
	}
	switch typed := value.(type) {
	case map[string]any:
		return []map[string]any{typed}, nil
	case []any:
		providers := make([]map[string]any, 0, len(typed))
		for index, item := range typed {
			fields, ok := item.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("item %d must be an object", index)
			}
			providers = append(providers, fields)
		}
		return providers, nil
	default:
		return nil, errors.New("invalid provider JSON")
	}
}

// Single quotes are accepted for compatibility with existing image configuration.
func decodeJSON(raw string) (any, error) {
	value, err := bootstrap.DecodeJSON(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid provider JSON: %w", err)
	}
	return value, nil
}

func secretConfigured(env bootstrap.Environment, name string) bool {
	_, configured := env[name]
	return configured
}

func hasSecrets(env bootstrap.Environment, name string) bool {
	for _, secret := range providerSecretDefinitions {
		variable := name + secret.Suffix
		if secret.Kind == bootstrap.ProviderValueSecret {
			if secretConfigured(env, variable) {
				return true
			}
			continue
		}
		location := ""
		if secret.ManagedLocation {
			location = env[variable+"LOCATION"]
		}
		if env[variable] != "" || env[variable+"FILE"] != "" || location != "" {
			return true
		}
	}
	return false
}

// Separate credential variables override credentials in JSON.
func applySecrets(env bootstrap.Environment, name string, fields map[string]any) error {
	for _, secret := range providerSecretDefinitions {
		variable := name + secret.Suffix
		if secret.Kind == bootstrap.ProviderValueSecret {
			if secretConfigured(env, variable) {
				fields[secret.Field] = env[variable]
			}
			continue
		}
		fileVariable := variable + "FILE"
		path := env[fileVariable]
		location := ""
		if secret.ManagedLocation {
			location = env[variable+"LOCATION"]
		}
		if path == "" && location == "" {
			continue
		}
		if _, exists := fields[secret.Field]; exists {
			return fmt.Errorf("%s: %s cannot be combined with %s", name, secret.Field, fileVariable)
		}
		if secret.ManagedLocation {
			fields[secret.Field] = location
		} else {
			fields[secret.Field] = path
		}
	}
	return nil
}

func deleteSecrets(env bootstrap.Environment, name string) {
	for _, secret := range providerSecretDefinitions {
		variable := name + secret.Suffix
		delete(env, variable)
		if secret.Kind == bootstrap.ProviderValueSecret {
			delete(env, variable+"_FILE")
			continue
		}
		delete(env, variable+"FILE")
		delete(env, variable+"LOCATION")
	}
}

func validateFilePath(value any, field string) error {
	name, ok := value.(string)
	if !ok || name == "" {
		return fmt.Errorf("%s must be a non-empty file path", field)
	}
	if strings.ContainsAny(name, "\\\r\n\x00") {
		return fmt.Errorf("%s contains an invalid path character", field)
	}
	if _, err := bootstrap.ResolveFile(name, ""); err != nil {
		return fmt.Errorf("%s: %w", field, err)
	}
	return nil
}

func quoteOption(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	return `"` + value + `"`
}

func nativeOption(key string, value any) (string, error) {
	switch typed := value.(type) {
	case string:
		if strings.ContainsAny(typed, "\r\n") {
			return "", fmt.Errorf("%s contains a line break", key)
		}
		return key + "=" + quoteOption(typed), nil
	case float64:
		return key + "=" + strconv.FormatFloat(typed, 'f', -1, 64), nil
	case bool:
		if typed {
			return key + "=1", nil
		}
		return key + "=0", nil
	default:
		return "", fmt.Errorf("%s must be a string, number, or boolean", key)
	}
}
