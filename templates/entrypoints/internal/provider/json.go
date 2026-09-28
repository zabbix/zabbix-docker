package provider

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

// SecretField binds the suffix of a credential variable to the provider option
// it fills.
type SecretField struct {
	Suffix string
	Field  string
}

// ClickHouseSecretFields are the credentials that ClickHouse providers accept
// through separate variables instead of the provider JSON. Every suffix also
// has a _FILE variant holding the name of a file with the value.
var ClickHouseSecretFields = []SecretField{
	{Suffix: "_USERNAME", Field: "username"},
	{Suffix: "_PASSWORD", Field: "password"},
	{Suffix: "_SSL_KEY_PASSWORD", Field: "ssl_key_password"},
}

// tlsFiles are the options naming TLS material that the image reads from the
// directories it owns, shared by every provider that connects over TLS.
var tlsFiles = []string{"ssl_cert_file", "ssl_key_file"}

// isJSONObject reports whether the value is a provider object instead of a
// native configuration value.
func isJSONObject(raw string) bool {
	return strings.HasPrefix(strings.TrimSpace(raw), "{")
}

// decodeObject parses one provider object.
func decodeObject(raw string) (map[string]any, error) {
	value, err := decodeJSON(raw)
	if err != nil {
		return nil, err
	}
	fields, ok := value.(map[string]any)
	if !ok || fields == nil {
		return nil, fmt.Errorf("invalid provider JSON")
	}
	return fields, nil
}

// decodeList parses one provider object or an array of them. An empty value
// means that no provider is configured.
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
		return nil, fmt.Errorf("invalid provider JSON")
	}
}

// decodeJSON accepts both JSON and the single-quoted form that the PHP
// configuration helper also reads.
func decodeJSON(raw string) (any, error) {
	var value any
	if err := json.Unmarshal([]byte(raw), &value); err == nil {
		return value, nil
	}
	if err := json.Unmarshal([]byte(strings.ReplaceAll(raw, "'", `"`)), &value); err != nil {
		return nil, fmt.Errorf("invalid provider JSON: %w", err)
	}
	return value, nil
}

// secretConfigured reports whether the variable or its _FILE variant is set.
func secretConfigured(env bootstrap.Environment, name string) bool {
	return env[name] != "" || env[name+"_FILE"] != ""
}

// hasSecrets reports whether any credential variable of the provider is set.
func hasSecrets(env bootstrap.Environment, name string) bool {
	for _, secret := range ClickHouseSecretFields {
		if secretConfigured(env, name+secret.Suffix) {
			return true
		}
	}
	return false
}

// applySecrets replaces the credential options with the values of the separate
// credential variables.
func applySecrets(env bootstrap.Environment, name string, fields map[string]any) error {
	for _, secret := range ClickHouseSecretFields {
		variable := name + secret.Suffix
		if !secretConfigured(env, variable) {
			continue
		}
		if err := bootstrap.ResolveSecretEnv(env, variable); err != nil {
			return err
		}
		fields[secret.Field] = env[variable]
	}
	return nil
}

// deleteSecrets removes the credential variables of the provider so that they
// are not passed to the service.
func deleteSecrets(env bootstrap.Environment, name string) {
	for _, secret := range ClickHouseSecretFields {
		delete(env, name+secret.Suffix)
		delete(env, name+secret.Suffix+"_FILE")
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
		return key + "=" + fmt.Sprint(typed), nil
	case bool:
		if typed {
			return key + "=1", nil
		}
		return key + "=0", nil
	default:
		return "", fmt.Errorf("%s must be a string, number, or boolean", key)
	}
}
