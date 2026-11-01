package storageprovider

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

type secretField struct {
	Suffix string
	Field  string
}

var clickHouseSecretFields = []secretField{
	{Suffix: "_USERNAME", Field: "username"},
	{Suffix: "_PASSWORD", Field: "password"},
	{Suffix: "_SSL_KEY_PASSWORD", Field: "ssl_key_password"},
}

var tlsFiles = []string{"ssl_cert_file", "ssl_key_file"}

func isJSONObject(raw string) bool {
	return strings.HasPrefix(strings.TrimSpace(raw), "{")
}

func decodeObject(raw string) (map[string]any, error) {
	value, err := decodeJSON(raw)
	if err != nil {
		return nil, err
	}
	fields, ok := value.(map[string]any)
	if !ok || fields == nil {
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

func secretConfigured(env bootstrap.Environment, name string) bool {
	return env[name] != "" || env[name+"_FILE"] != ""
}

func hasSecrets(env bootstrap.Environment, name string) bool {
	for _, secret := range clickHouseSecretFields {
		if secretConfigured(env, name+secret.Suffix) {
			return true
		}
	}
	return false
}

// Separate credential variables override credentials in JSON.
func applySecrets(env bootstrap.Environment, name string, fields map[string]any) error {
	for _, secret := range clickHouseSecretFields {
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

func deleteSecrets(env bootstrap.Environment, name string) {
	for _, secret := range clickHouseSecretFields {
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
