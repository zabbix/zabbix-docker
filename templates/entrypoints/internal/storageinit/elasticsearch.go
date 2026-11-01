package storageinit

import (
	"context"
	"errors"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/elasticsearch"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/storageprovider"
)

type elasticsearchSchemaOptions struct {
	Create   *bool `json:"create"`
	Shards   int   `json:"shards"`
	Replicas int   `json:"replicas"`
}

func elasticsearchPreparation(config storageprovider.ProviderConfig) (preparation, error) {
	options := elasticsearchSchemaOptions{Shards: 5, Replicas: 1}
	if err := decodeOptions(config.SchemaOptions, &options); err != nil {
		return preparation{}, err
	}
	if options.Create == nil {
		return preparation{}, errors.New("_SCHEMA requires an explicit boolean create")
	}
	endpoint, ok := config.Fields["url"].(string)
	if !ok || endpoint == "" {
		return preparation{}, errors.New("url must be a non-empty string")
	}
	dateIndex, err := providerBool(config.Fields, "date_index", false)
	if err != nil {
		return preparation{}, err
	}
	schema := elasticsearch.Schema{
		ValueTypes: config.ValueTypes, DateIndex: dateIndex,
		Shards: options.Shards, Replicas: options.Replicas,
	}
	if err := schema.Validate(); err != nil {
		return preparation{}, err
	}
	client, err := elasticsearch.NewClient(endpoint)
	if err != nil {
		return preparation{}, err
	}

	return preparation{run: func(ctx context.Context) error {
		defer client.Close()
		bootstrap.LogInfo("** Preparing Elasticsearch schema for %s (url=%q, date_index=%t, create=%t)",
			config.Name, client.Endpoint(), schema.DateIndex, *options.Create)
		if err := client.Prepare(ctx, schema, *options.Create); err != nil {
			return err
		}
		bootstrap.LogInfo("** Elasticsearch schema ready for %s", config.Name)
		return nil
	}}, nil
}
