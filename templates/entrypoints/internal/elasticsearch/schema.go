// Package elasticsearch prepares history indices, templates and ingest pipelines.
package elasticsearch

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// Mappings follow Zabbix database/elasticsearch/elasticsearch.map.
//
//go:embed schemas/*.json
var schemas embed.FS

type Schema struct {
	ValueTypes       []string // Validated by storageinit.
	DateIndex        bool
	Shards, Replicas int
}

type object struct {
	path, method, alternate, createPath string
	body                                []byte
}

func (s Schema) Validate() error {
	if s.Shards < 1 || s.Replicas < 0 {
		return errors.New("shards must be positive and replicas non-negative")
	}
	return nil
}

func (s Schema) objects() ([]object, error) {
	var objects []object
	for _, kind := range s.ValueTypes {
		if s.DateIndex || kind == "json" {
			pipeline, err := s.pipelineObject(kind)
			if err != nil {
				return nil, err
			}
			objects = append(objects, pipeline)
		}
		index, err := s.indexObject(kind)
		if err != nil {
			return nil, err
		}
		objects = append(objects, index)
	}
	return objects, nil
}

func (s Schema) pipelineObject(kind string) (object, error) {
	pipelineName := "json"
	pipeline := struct {
		Processors []any `json:"processors"`
	}{}
	if kind == "json" {
		// Unlike the upstream mapping file, arrays and scalars also go to value_raw;
		// removing value_parsed prevents a conflict with its object mapping.
		data, err := schemas.ReadFile("schemas/json_pipeline.json")
		if err != nil {
			return object{}, err
		}
		if err := json.Unmarshal(data, &pipeline); err != nil {
			return object{}, err
		}
	}
	if s.DateIndex {
		pipelineName = kind + "-pipeline"
		pipeline.Processors = append(pipeline.Processors, map[string]any{
			"date_index_name": map[string]any{
				"field": "clock", "date_formats": []string{"UNIX"},
				"index_name_prefix": kind + "-", "date_rounding": "d",
			},
		})
	}
	body, err := json.Marshal(pipeline)
	if err != nil {
		return object{}, err
	}
	return object{path: "/_ingest/pipeline/" + pipelineName, method: http.MethodGet, body: body}, nil
}

func (s Schema) indexObject(kind string) (object, error) {
	mapping, err := schemas.ReadFile("schemas/" + kind + ".json")
	if err != nil {
		return object{}, err
	}
	settings := map[string]any{"number_of_shards": s.Shards, "number_of_replicas": s.Replicas}
	if !s.DateIndex && kind == "json" {
		settings["default_pipeline"] = "json"
	}
	definition := map[string]any{"settings": settings, "mappings": json.RawMessage(mapping)}
	obj := object{path: "/" + kind, method: http.MethodHead}
	if s.DateIndex {
		obj.path = "/_index_template/" + kind + "_template"
		obj.alternate = "/_template/" + kind + "_template"
		obj.createPath = obj.path + "?create=true"
		definition = map[string]any{"index_patterns": []string{kind + "-*"}, "template": definition}
	}
	obj.body, err = json.Marshal(definition)
	if err != nil {
		return object{}, err
	}
	return obj, nil
}

func (c *Client) objectExists(ctx context.Context, obj object) (bool, error) {
	exists, err := c.exists(ctx, obj.method, obj.path)
	if err != nil || exists || obj.alternate == "" {
		return exists, err
	}
	// A legacy template from the Zabbix documentation also counts as existing.
	return c.exists(ctx, http.MethodHead, obj.alternate)
}

// Prepare checks for existing objects and optionally creates missing ones.
func (c *Client) Prepare(ctx context.Context, schema Schema, create bool) error {
	objects, err := schema.objects()
	if err != nil {
		return err
	}

	for _, obj := range objects {
		exists, err := c.objectExists(ctx, obj)
		if err != nil {
			return fmt.Errorf("check %s: %w", obj.path, err)
		}
		if exists {
			continue
		}

		if !create {
			return fmt.Errorf("%s is missing (creation is disabled)", obj.path)
		}
		path := obj.createPath
		if path == "" {
			path = obj.path
		}
		createErr := c.request(ctx, http.MethodPut, path, obj.body)
		if createErr != nil {
			createErr = fmt.Errorf("create %s: %w", obj.path, createErr)
		}

		exists, err = c.objectExists(ctx, obj)
		if err != nil {
			return errors.Join(createErr, fmt.Errorf("check after creation %s: %w", obj.path, err))
		}
		if !exists {
			if createErr != nil {
				return createErr
			}
			return fmt.Errorf("%s is still missing after creation", obj.path)
		}
	}
	return nil
}
