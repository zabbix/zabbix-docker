package storageinit

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"path/filepath"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/clickhouse"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/storageprovider"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/vault"
)

type clickHouseSchemaOptions struct {
	Create     *bool  `json:"create"`
	Cluster    string `json:"cluster"`
	Replicated bool   `json:"replicated"`
	TTL        int64  `json:"ttl"`
	Partition  string `json:"partition"`
}

func clickHousePreparation(ctx context.Context, env bootstrap.Environment, dataset string, config storageprovider.ProviderConfig) (preparation, error) {
	options := clickHouseSchemaOptions{TTL: 2678400, Partition: "toDate"}
	if err := decodeOptions(config.SchemaOptions, &options); err != nil {
		return preparation{}, err
	}
	if options.Create == nil {
		return preparation{}, errors.New("_SCHEMA requires an explicit boolean create")
	}

	database, ok := config.Fields["db"].(string)
	if config.Fields["db"] != nil && !ok {
		return preparation{}, fmt.Errorf("db must be a string, got %T", config.Fields["db"])
	}

	schema := clickhouse.Schema{
		Database: database, Dataset: dataset, ValueTypes: config.ValueTypes,
		Cluster: options.Cluster, Replicated: options.Replicated,
		TTL: options.TTL, Partition: options.Partition,
	}
	if err := schema.Validate(); err != nil {
		return preparation{}, err
	}

	conn, err := clickHouseConnection(ctx, env, dataset, config.Fields)
	if err != nil {
		return preparation{}, err
	}
	client, err := clickhouse.NewClient(ctx, conn)
	if err != nil {
		return preparation{}, err
	}

	return preparation{run: func(ctx context.Context) error {
		defer client.Close()
		cluster := ""
		if schema.Cluster != "" {
			cluster = fmt.Sprintf(", cluster=%q", schema.Cluster)
		}
		bootstrap.LogInfo("** Preparing ClickHouse schema for %s (url=%q, database=%q%s, create=%t)",
			config.Name, conn.URL, schema.Database, cluster, *options.Create)
		if err := client.Prepare(ctx, schema, *options.Create); err != nil {
			return err
		}
		bootstrap.LogInfo("** ClickHouse schema ready for %s", config.Name)
		return nil
	}}, nil
}

func clickHouseConnection(ctx context.Context, env bootstrap.Environment, dataset string, fields map[string]any) (clickhouse.Connection, error) {
	values := make(map[string]string)
	keys := []string{
		"url", "username", "password", "source_ip",
		"ssl_cert_file", "ssl_key_file", "ssl_key_password",
	}
	if dataset == "otel" {
		keys = append(keys, "vault_path")
	}
	for _, key := range keys {
		if value := fields[key]; value != nil {
			text, ok := value.(string)
			if !ok {
				return clickhouse.Connection{}, fmt.Errorf("%s must be a string, got %T", key, value)
			}
			values[key] = text
		}
	}

	conn := clickhouse.Connection{
		URL: values["url"], Username: values["username"], Password: values["password"],
		SourceIP: values["source_ip"], CertFile: values["ssl_cert_file"],
		KeyFile: values["ssl_key_file"], KeyPassword: values["ssl_key_password"],
	}
	var err error
	if conn.SourceIP == "" {
		conn.SourceIP = env["ZBX_SOURCEIP"]
	}

	conn.VerifyPeer, err = providerBool(fields, "ssl_verify_peer", true)
	if err != nil {
		return clickhouse.Connection{}, err
	}
	conn.VerifyHost, err = providerBool(fields, "ssl_verify_host", true)
	if err != nil {
		return clickhouse.Connection{}, err
	}

	conn.CADirectory = env["ZBX_SSLCALOCATION"]
	conn.CertFile, err = resolveTLSFile(env, conn.CertFile, "ZBX_SSLCERTLOCATION", "certs")
	if err != nil {
		return clickhouse.Connection{}, err
	}
	conn.KeyFile, err = resolveTLSFile(env, conn.KeyFile, "ZBX_SSLKEYLOCATION", "keys")
	if err != nil {
		return clickhouse.Connection{}, err
	}

	if vaultPath := values["vault_path"]; vaultPath != "" {
		if conn.Username != "" || conn.Password != "" {
			return clickhouse.Connection{}, errors.New("vault_path cannot be combined with username/password for schema preparation")
		}
		vaultEnv := maps.Clone(env)
		vaultEnv["ZBX_VAULTDBPATH"] = vaultPath
		creds, err := vault.Fetch(ctx, vaultEnv)
		if err != nil {
			return clickhouse.Connection{}, err
		}
		conn.Username, conn.Password = creds.Username, creds.Password
	}
	return conn, nil
}

func resolveTLSFile(env bootstrap.Environment, path, location, subdir string) (string, error) {
	if path == "" {
		return "", nil
	}
	dir := env[location]
	if dir == "" && env["ZABBIX_USER_HOME_DIR"] != "" {
		dir = filepath.Join(env["ZABBIX_USER_HOME_DIR"], "ssl", subdir)
	}
	if dir == "" && !filepath.IsAbs(path) {
		return "", fmt.Errorf("%s is required for relative TLS files", location)
	}
	return bootstrap.ResolveFile(path, dir)
}
