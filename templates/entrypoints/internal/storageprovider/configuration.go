package storageprovider

import (
	"fmt"
	"strings"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

const schemaSuffix = "_SCHEMA"

// ProviderConfig contains a decoded provider and its schema options.
type ProviderConfig struct {
	Name, Provider, SchemaOptions string
	Fields                        map[string]any
	ValueTypes                    []string
	definition                    providerDefinition
}

// Configuration caches decoded providers so preparation and rendering resolve secrets once.
type Configuration struct {
	env     bootstrap.Environment
	param   *Parameter
	names   []string
	decoded map[string]ProviderConfig
}

// NewConfiguration validates provider indices. JSON and secrets are decoded on demand.
func NewConfiguration(env bootstrap.Environment, param *Parameter) (*Configuration, error) {
	names, err := param.indexedNames(env)
	if err != nil {
		return nil, err
	}
	return &Configuration{env: env, param: param, names: names, decoded: make(map[string]ProviderConfig)}, nil
}

// ConfigsToPrepare decodes only providers with a non-empty _SCHEMA.
func (c *Configuration) ConfigsToPrepare() ([]ProviderConfig, error) {
	var configs []ProviderConfig
	for _, name := range c.names {
		if strings.TrimSpace(c.env[name+schemaSuffix]) == "" {
			continue
		}
		if !isJSONObject(c.env[name]) {
			return nil, fmt.Errorf("%s requires a JSON provider", name+schemaSuffix)
		}
		config, err := c.decode(name)
		if err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}
	return configs, nil
}

func (c *Configuration) decode(name string) (ProviderConfig, error) {
	if config, ok := c.decoded[name]; ok {
		return config, nil
	}
	config, err := c.param.decode(c.env, name)
	if err != nil {
		return ProviderConfig{}, err
	}
	config.SchemaOptions = c.env[name+schemaSuffix]
	c.decoded[name] = config
	return config, nil
}
