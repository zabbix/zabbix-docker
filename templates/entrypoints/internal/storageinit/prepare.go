// Package storageinit prepares databases for history and telemetry providers.
package storageinit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"path/filepath"
	"strings"
	"time"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/storageprovider"
)

const (
	historyDataset   = "history"
	telemetryDataset = "otel"
)

// PrepareHistory initializes the selected history providers without configuring Zabbix.
func PrepareHistory(env bootstrap.Environment) (int, error) {
	return prepareProviders(env, storageprovider.History, historyDataset)
}

// PrepareTelemetry initializes the selected telemetry provider without configuring Zabbix.
func PrepareTelemetry(env bootstrap.Environment) (int, error) {
	return prepareProviders(env, storageprovider.Telemetry, telemetryDataset)
}

// ConfigureHistory prepares selected schemas and writes native history configuration.
func ConfigureHistory(env bootstrap.Environment) error {
	return configureProviders(env, storageprovider.History, historyDataset)
}

// ConfigureTelemetry prepares selected schemas and writes native telemetry configuration.
func ConfigureTelemetry(env bootstrap.Environment) error {
	return configureProviders(env, storageprovider.Telemetry, telemetryDataset)
}

func prepareProviders(env bootstrap.Environment, param *storageprovider.Parameter, dataset string) (int, error) {
	configuration, err := storageprovider.NewConfiguration(env, param)
	if err != nil {
		return 0, err
	}
	configs, err := configuration.ConfigsToPrepare()
	if err != nil {
		return 0, err
	}
	if len(configs) != 0 {
		if err := bootstrap.PrepareCertDir(filepath.Join(env["ZABBIX_USER_HOME_DIR"], "ssl", "ssl_ca"), env["ZBX_SSLCALOCATION"]); err != nil {
			return 0, err
		}
	}
	return prepareSchemas(env, configs, dataset)
}

func configureProviders(env bootstrap.Environment, param *storageprovider.Parameter, dataset string) error {
	configuration, err := storageprovider.NewConfiguration(env, param)
	if err != nil {
		return err
	}
	configs, err := configuration.ConfigsToPrepare()
	if err != nil {
		return err
	}
	if err := configuration.ConfigureNative(); err != nil {
		return err
	}
	_, err = prepareSchemas(env, configs, dataset)
	return err
}

type preparation struct {
	name string
	run  func(context.Context) error
}

func (p preparation) prepare(ctx context.Context) error {
	const interval = 10 * time.Second
	return bootstrap.Retry(ctx, bootstrap.RetryOptions{
		Attempts: 3,
		Interval: interval,
		OnRetry: func(err error) {
			bootstrap.LogWarn("** %s: %v; retrying in %s", p.name, err, interval)
		},
	}, func() error {
		err := p.run(ctx)
		var connectionError *net.OpError
		if err != nil && (!errors.As(err, &connectionError) || connectionError.Op != "dial") {
			return bootstrap.Stop(err)
		}
		return err
	})
}

func prepareSchemas(env bootstrap.Environment, configs []storageprovider.ProviderConfig, dataset string) (int, error) {
	ctx, stop := bootstrap.TerminationContext()
	defer stop()

	var preparations []preparation
	for _, config := range configs {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		if dataset == historyDataset {
			if err := validateHistoryTypes(config.ValueTypes); err != nil {
				return 0, fmt.Errorf("%s: %w", config.Name, err)
			}
		}

		var task preparation
		var err error
		switch config.Provider {
		case "clickhouse":
			task, err = clickHousePreparation(ctx, env, dataset, config)
		case "elasticsearch":
			task, err = elasticsearchPreparation(config)
		default:
			err = fmt.Errorf("schema preparation is not supported for %s", config.Provider)
		}
		if err != nil {
			return 0, fmt.Errorf("%s: %w", config.Name, err)
		}
		task.name = config.Name
		preparations = append(preparations, task)
	}

	for _, task := range preparations {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		if err := task.prepare(ctx); err != nil {
			return 0, fmt.Errorf("%s: %w", task.name, err)
		}
	}
	return len(preparations), nil
}

func decodeOptions(raw string, options any) error {
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(options); err != nil {
		return fmt.Errorf("invalid _SCHEMA: %w", err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return errors.New("_SCHEMA must contain one JSON object")
	}
	return nil
}

func validateHistoryTypes(types []string) error {
	if len(types) == 0 {
		return errors.New("history schema requires value types")
	}
	for _, kind := range types {
		switch kind {
		case "dbl", "uint", "str", "text", "log", "json":
		default:
			return fmt.Errorf("unsupported history type %q", kind)
		}
	}
	return nil
}

func providerBool(fields map[string]any, key string, fallback bool) (bool, error) {
	value, exists := fields[key]
	if !exists {
		return fallback, nil
	}
	switch v := value.(type) {
	case bool:
		return v, nil
	case float64:
		if v == 0 || v == 1 {
			return v == 1, nil
		}
	}
	return false, fmt.Errorf("%s must be boolean or 0/1", key)
}
