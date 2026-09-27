package provider

import (
	"fmt"
	"sort"
	"strings"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

// valueTypeOption is the JSON option holding history value types. The native
// configuration calls the same option value_types.
const valueTypeOption = "types"

// schema describes how the options of one provider are rendered. Option value
// ranges, mandatory options and their combinations are validated by Zabbix.
type schema struct {
	// valueTypes reports that the provider takes the types array, rendered as
	// the native value_types option.
	valueTypes bool
	// files are the options naming a file that the image reads from a directory
	// it owns, so only a file name without a path is accepted.
	files []string
	// web are the options the frontend configuration understands.
	web []string
	// webDefaults are emitted for the frontend even when the provider does not
	// set them, because the frontend reads its options without a default.
	webDefaults map[string]any
	// secrets reports that the credential variables listed in
	// ClickHouseSecretFields may fill options of this provider.
	secrets bool
}

// parameter describes one Zabbix configuration parameter that is configured
// from indexed JSON environment variables.
type parameter struct {
	// Prefix is the variable prefix of one provider, without the underscore
	// that precedes the index.
	Prefix string
	// Plural is the frontend variable holding a JSON array of providers.
	Plural string
	// pluralInput reports that Plural is also accepted as input.
	pluralInput bool
	// maxProviders limits the number of indexed providers; zero means no limit.
	maxProviders int
	// providers are the providers the image knows by name.
	providers map[string]schema
	// providerError describes the accepted provider names.
	providerError string
	// reserved options are configured by the image itself and are rejected with
	// their own explanation.
	reserved map[string]string
	// defaults applies the connection defaults of the image after the generic
	// checks.
	defaults func(fields map[string]any)
}

// managedLocations are the options pointing into directories the image owns:
// it prepares the certificate and CA directories itself, so a provider cannot
// redirect them. Every other option is passed to the service, which decides
// whether it is valid.
var managedLocations = reservedOptions("is managed by the image and must not be set",
	"ssl_ca_file", "ssl_ca_location", "ssl_cert_location", "ssl_key_location")

// reservedOptions explains every listed option with the same reason.
func reservedOptions(reason string, keys ...string) map[string]string {
	options := make(map[string]string, len(keys))
	for _, key := range keys {
		options[key] = key + " " + reason
	}
	return options
}

// ConfigureNative replaces every JSON value of param with its native
// configuration value. Values already written in the native format are kept
// unchanged, so both formats can be mixed.
func ConfigureNative(env bootstrap.Environment, param *parameter) error {
	names, err := param.indexedNames(env)
	if err != nil || len(names) == 0 {
		return err
	}

	values := make(map[string]string, len(names))
	for _, name := range names {
		if !isJSONObject(env[name]) {
			if hasSecrets(env, name) {
				return fmt.Errorf("%s credentials require a JSON provider", name)
			}
			continue
		}
		providerName, current, fields, err := param.decode(env, name)
		if err != nil {
			return err
		}
		value, err := renderNative(providerName, current, fields)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		values[name] = value
	}

	for name, value := range values {
		env[name] = value
		deleteSecrets(env, name)
	}
	return nil
}

// WebProviders returns the providers of param as frontend configuration
// objects and removes the input variables from the environment. Credentials
// supplied through separate variables take precedence over JSON fields.
func WebProviders(env bootstrap.Environment, param *parameter) ([]map[string]any, error) {
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
		_, current, fields, err := param.decode(env, name)
		if err != nil {
			return nil, err
		}
		providers = append(providers, webFields(env, name, current, fields))
	}
	for _, fields := range plural {
		_, current, err := param.validate(env, param.Plural, fields)
		if err != nil {
			return nil, err
		}
		providers = append(providers, webFields(env, param.Plural, current, fields))
	}

	for _, name := range names {
		delete(env, name)
		deleteSecrets(env, name)
	}
	deleteSecrets(env, param.Plural)
	return providers, nil
}

// indexedNames returns the consecutive provider variables of param.
func (p *parameter) indexedNames(env bootstrap.Environment) ([]string, error) {
	names, err := IndexedNames(env, p.Prefix)
	if err != nil {
		return nil, err
	}
	if p.maxProviders != 0 && len(names) > p.maxProviders {
		return nil, fmt.Errorf("%s is not supported; use %s", bootstrap.IndexedName(p.Prefix, p.maxProviders),
			bootstrap.IndexedName(p.Prefix, 0))
	}
	return names, nil
}

// pluralProviders decodes the frontend variable of param. indexed is the
// number of indexed providers, which cannot be combined with it.
func (p *parameter) pluralProviders(env bootstrap.Environment, indexed int) ([]map[string]any, error) {
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

// decode reads and validates the provider stored in the named variable.
func (p *parameter) decode(env bootstrap.Environment, name string) (string, schema, map[string]any, error) {
	fields, err := decodeObject(env[name])
	if err != nil {
		return "", schema{}, nil, fmt.Errorf("%s: %w", name, err)
	}
	providerName, current, err := p.validate(env, name, fields)
	if err != nil {
		return "", schema{}, nil, err
	}
	return providerName, current, fields, nil
}

// validate checks what the image itself has to know about the provider: the
// option names it generates, the files it owns, and the credentials it
// resolves. It then applies the credentials and the image defaults.
func (p *parameter) validate(env bootstrap.Environment, name string, fields map[string]any) (string, schema, error) {
	providerName, ok := fields["provider"].(string)
	if !ok || providerName == "" {
		return "", schema{}, fmt.Errorf("%s: %s", name, p.providerError)
	}
	current, known := p.providers[providerName]
	if !known {
		return "", schema{}, fmt.Errorf("%s: %s", name, p.providerError)
	}

	for key, value := range fields {
		if key == "" || strings.ContainsAny(key, ",=;\r\n") {
			return "", schema{}, fmt.Errorf("%s: invalid option %q", name, key)
		}
		if reason, isReserved := p.reserved[key]; isReserved {
			return "", schema{}, fmt.Errorf("%s: %s", name, reason)
		}
		if key == valueTypeOption && current.valueTypes {
			if err := validateValueTypes(value); err != nil {
				return "", schema{}, fmt.Errorf("%s: %w", name, err)
			}
			continue
		}
		if current.isFile(key) {
			if err := validateFileName(value, key); err != nil {
				return "", schema{}, fmt.Errorf("%s: %w", name, err)
			}
		}
	}

	if !current.secrets {
		if hasSecrets(env, name) {
			return "", schema{}, fmt.Errorf("%s credentials are only supported by ClickHouse providers", name)
		}
	} else if err := applySecrets(env, name, fields); err != nil {
		return "", schema{}, err
	}

	if p.defaults != nil {
		p.defaults(fields)
	}
	return providerName, current, nil
}

// leadingOptions are rendered before the others because they identify the
// connection. Every remaining option follows in alphabetical order, which
// keeps the generated value stable without the image knowing the option.
var leadingOptions = []string{"url", "db"}

// renderNative formats fields as one native configuration value. Options the
// image does not know are rendered as well, so that the service can accept or
// reject them.
func renderNative(providerName string, current schema, fields map[string]any) (string, error) {
	options := make([]string, 0, len(fields))
	rendered := map[string]bool{"provider": true}
	if current.valueTypes {
		rendered[valueTypeOption] = true
		if value, exists := fields[valueTypeOption]; exists {
			options = append(options, "value_types="+quoteOption(joinValueTypes(value)))
		}
	}

	order := make([]string, 0, len(fields))
	for _, key := range leadingOptions {
		order = append(order, key)
		rendered[key] = true
	}
	remaining := make([]string, 0, len(fields))
	for key := range fields {
		if !rendered[key] {
			remaining = append(remaining, key)
		}
	}
	sort.Strings(remaining)

	for _, key := range append(order, remaining...) {
		value, exists := fields[key]
		if !exists {
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

// webFields copies the options the frontend understands, on top of the defaults
// it needs to receive explicitly. The options left out are reported once,
// because a typo is otherwise invisible.
func webFields(env bootstrap.Environment, name string, current schema, fields map[string]any) map[string]any {
	result := make(map[string]any, len(current.web)+len(current.webDefaults))
	for key, value := range current.webDefaults {
		result[key] = value
	}
	for _, key := range current.web {
		if value, exists := fields[key]; exists {
			result[key] = value
		}
	}
	dropped := make([]string, 0, len(fields))
	for key := range fields {
		if _, used := result[key]; !used {
			dropped = append(dropped, key)
		}
	}
	if len(dropped) != 0 {
		sort.Strings(dropped)
		bootstrap.LogDebug(env, "** %s: options not used by the frontend: %s", name, strings.Join(dropped, ", "))
	}
	return result
}

// isFile reports whether the option names a file owned by the image.
func (s schema) isFile(key string) bool {
	for _, file := range s.files {
		if file == key {
			return true
		}
	}
	return false
}

// validateValueTypes checks only that the value types can be rendered as one
// native value_types option; the service validates the names themselves.
func validateValueTypes(value any) error {
	items, ok := value.([]any)
	if !ok || len(items) == 0 {
		return fmt.Errorf("%s must be a non-empty array", valueTypeOption)
	}
	for _, item := range items {
		valueType, ok := item.(string)
		if !ok || valueType == "" || strings.ContainsAny(valueType, "\r\n") {
			return fmt.Errorf("%s must contain value type names", valueTypeOption)
		}
	}
	return nil
}

func joinValueTypes(value any) string {
	items, _ := value.([]any)
	valueTypes := make([]string, 0, len(items))
	for _, item := range items {
		if valueType, ok := item.(string); ok {
			valueTypes = append(valueTypes, valueType)
		}
	}
	return strings.Join(valueTypes, ",")
}
