package storageprovider

import (
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"strings"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

// valueTypeOption is the JSON option holding history value types. The native
// configuration calls the same option value_types.
const valueTypeOption = "types"

type providerDefinition struct {
	valueTypes  bool
	files       []string
	web         []string
	webDefaults map[string]any // Values required by PHP when the JSON option is omitted.
	secrets     bool
}

// Parameter describes the environment variables and options of a provider family.
type Parameter struct {
	Prefix        string
	Plural        string
	pluralInput   bool // Accept an unindexed JSON array from Plural.
	maxProviders  int  // Zero means no limit.
	providers     map[string]providerDefinition
	providerError string
	reserved      map[string]string
	defaults      func(fields map[string]any)
}

var managedLocations = []string{"ssl_ca_file", "ssl_ca_location", "ssl_cert_location", "ssl_key_location"}

// ConfigureNative converts JSON providers and removes their credential and _SCHEMA variables.
func (c *Configuration) ConfigureNative() error {
	values := make(map[string]string, len(c.names))
	for _, name := range c.names {
		if !looksLikeJSONObject(c.env[name]) {
			if hasSecrets(c.env, name) {
				return fmt.Errorf("%s credentials require a JSON provider", name)
			}
			continue
		}
		config, err := c.decode(name)
		if err != nil {
			return err
		}
		value, err := renderNative(config.Provider, config.definition, maps.Clone(config.Fields))
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		values[name] = value
	}

	for name, value := range values {
		c.env[name] = value
		deleteSecrets(c.env, name)
	}
	for _, name := range c.names {
		delete(c.env, name+schemaSuffix)
	}
	return nil
}

// WebProviders builds frontend configuration and removes its input variables.
func WebProviders(env bootstrap.Environment, param *Parameter) ([]map[string]any, error) {
	names, err := param.indexedNames(env)
	if err != nil {
		return nil, err
	}
	plural, err := param.pluralProviders(env, len(names))
	if err != nil {
		return nil, err
	}

	providers := make([]map[string]any, 0, len(names)+len(plural))
	for _, name := range names {
		config, err := param.decode(env, name)
		if err != nil {
			return nil, err
		}
		if len(config.definition.web) != 0 {
			providers = append(providers, webFields(config.definition, config.Fields))
		}
	}
	for _, fields := range plural {
		config, err := param.validate(env, param.Plural, fields)
		if err != nil {
			return nil, err
		}
		if len(config.definition.web) != 0 {
			providers = append(providers, webFields(config.definition, config.Fields))
		}
	}

	for _, name := range names {
		delete(env, name)
		deleteSecrets(env, name)
		delete(env, name+schemaSuffix)
	}
	deleteSecrets(env, param.Plural)
	return providers, nil
}

func (p *Parameter) indexedNames(env bootstrap.Environment) ([]string, error) {
	names, err := indexedNames(env, p.Prefix)
	if err != nil {
		return nil, err
	}
	if p.maxProviders != 0 && len(names) > p.maxProviders {
		return nil, fmt.Errorf("%s is not supported; use %s", bootstrap.IndexedName(p.Prefix, p.maxProviders),
			bootstrap.IndexedName(p.Prefix, 0))
	}
	return names, nil
}

func (p *Parameter) pluralProviders(env bootstrap.Environment, indexed int) ([]map[string]any, error) {
	raw := env[p.Plural]
	if !p.pluralInput {
		if strings.TrimSpace(raw) != "" {
			return nil, fmt.Errorf("%s is not supported; use %s", p.Plural, bootstrap.IndexedName(p.Prefix, 0))
		}
		return nil, nil
	}
	providers, err := decodeList(raw)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", p.Plural, err)
	}
	if len(providers) != 0 && indexed != 0 {
		return nil, fmt.Errorf("cannot combine %s and indexed %s_N variables", p.Plural, p.Prefix)
	}
	if hasSecrets(env, p.Plural) && len(providers) != 1 {
		return nil, fmt.Errorf("%s credentials require one unindexed JSON provider", p.Plural)
	}
	return providers, nil
}

func (p *Parameter) decode(env bootstrap.Environment, name string) (ProviderConfig, error) {
	fields, err := decodeObject(env[name])
	if err != nil {
		return ProviderConfig{}, fmt.Errorf("%s: %w", name, err)
	}
	return p.validate(env, name, fields)
}

func (p *Parameter) validate(env bootstrap.Environment, name string, fields map[string]any) (ProviderConfig, error) {
	providerName, _ := fields["provider"].(string)
	current, known := p.providers[providerName]
	if !known {
		return ProviderConfig{}, fmt.Errorf("%s: %s", name, p.providerError)
	}

	for key, value := range fields {
		if key == "" || strings.ContainsAny(key, ",=;\r\n") {
			return ProviderConfig{}, fmt.Errorf("%s: invalid option %q", name, key)
		}
		if slices.Contains(managedLocations, key) {
			return ProviderConfig{}, fmt.Errorf("%s: %s is managed by the image and must not be set", name, key)
		}
		if reason, isReserved := p.reserved[key]; isReserved {
			return ProviderConfig{}, fmt.Errorf("%s: %s", name, reason)
		}
		if key == valueTypeOption && current.valueTypes {
			types, err := parseValueTypes(value)
			if err != nil {
				return ProviderConfig{}, fmt.Errorf("%s: %w", name, err)
			}
			fields[key] = types
			continue
		}
		if slices.Contains(current.files, key) {
			if err := validateFilePath(value, key); err != nil {
				return ProviderConfig{}, fmt.Errorf("%s: %w", name, err)
			}
			bootstrap.LogWarn("%s field %s is deprecated; use %s_%sFILE", name, key, name,
				strings.ToUpper(strings.TrimSuffix(key, "_file")))
		}
	}

	if !current.secrets {
		if hasSecrets(env, name) {
			return ProviderConfig{}, fmt.Errorf("%s credentials are only supported by ClickHouse providers", name)
		}
	} else if err := applySecrets(env, name, fields); err != nil {
		return ProviderConfig{}, err
	}

	if p.defaults != nil {
		p.defaults(fields)
	}
	config := ProviderConfig{Name: name, Provider: providerName, Fields: fields, definition: current}
	if current.valueTypes {
		config.ValueTypes, _ = fields[valueTypeOption].([]string)
	}
	return config, nil
}

func renderNative(providerName string, current providerDefinition, fields map[string]any) (string, error) {
	// Zabbix joins each TLS location and file name, even for absolute paths.
	for _, key := range current.files {
		if value, _ := fields[key].(string); filepath.IsAbs(value) {
			fields[strings.TrimSuffix(key, "_file")+"_location"] = filepath.Dir(value)
			fields[key] = filepath.Base(value)
		}
	}

	options := make([]string, 0, len(fields))
	for _, key := range slices.Sorted(maps.Keys(fields)) {
		if key == "provider" {
			continue
		}
		value := fields[key]
		if key == valueTypeOption && current.valueTypes {
			options = append(options, "value_types="+quoteOption(strings.Join(value.([]string), ",")))
			continue
		}
		option, err := nativeOption(key, value)
		if err != nil {
			return "", err
		}
		options = append(options, option)
	}
	return providerName + ";" + strings.Join(options, ","), nil
}

func webFields(current providerDefinition, fields map[string]any) map[string]any {
	result := make(map[string]any, len(current.web)+len(current.webDefaults))
	for key, value := range current.webDefaults {
		result[key] = value
	}
	for _, key := range current.web {
		if value, exists := fields[key]; exists {
			result[key] = value
		}
	}
	return result
}

func parseValueTypes(value any) ([]string, error) {
	items, ok := value.([]any)
	if !ok || len(items) == 0 {
		return nil, fmt.Errorf("%s must be a non-empty array", valueTypeOption)
	}
	types := make([]string, 0, len(items))
	for _, item := range items {
		valueType, ok := item.(string)
		if !ok || valueType == "" || strings.ContainsAny(valueType, "\r\n") {
			return nil, fmt.Errorf("%s must contain value type names", valueTypeOption)
		}
		if !slices.Contains(types, valueType) {
			types = append(types, valueType)
		}
	}
	return types, nil
}
