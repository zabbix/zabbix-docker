// Package clickhouse prepares Zabbix history and OpenTelemetry schemas.
package clickhouse

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

//go:embed schemas/*.sql
var schemas embed.FS

type Schema struct {
	Database   string
	Dataset    string
	ValueTypes []string // Validated by storageinit.
	Cluster    string
	Replicated bool
	TTL        int64 // Seconds; zero disables TTL.
	Partition  string
}

type table struct {
	name, columns, timestamp, orderBy, shardKey string
}

func (s Schema) Validate() error {
	if s.Database == "" {
		return errors.New("database must be non-empty")
	}

	if s.TTL < 0 {
		return errors.New("ttl must be non-negative seconds")
	}

	switch s.Partition {
	case "toDate", "toYYYYMM":
	default:
		return errors.New("partition must be toDate or toYYYYMM")
	}

	return nil
}

func (s Schema) tables() ([]table, error) {
	var tables []table
	switch s.Dataset {
	case "history":
		for _, valueType := range s.ValueTypes {
			name := "history_" + valueType
			if valueType == "dbl" {
				name = "history"
			}

			columns, err := schemas.ReadFile("schemas/" + name + ".sql")
			if err != nil {
				return nil, err
			}

			tables = append(tables, table{name, string(columns),
				"clock_ns", "(itemid, clock_ns)", "intHash64(itemid)"})
		}
	case "otel":
		for _, name := range []string{"otel_logs", "otel_traces", "otel_metrics_gauge", "otel_metrics_sum",
			"otel_metrics_histogram", "otel_metrics_exponential_histogram"} {
			columns, err := schemas.ReadFile("schemas/" + name + ".sql")
			if err != nil {
				return nil, err
			}

			t := table{name, string(columns), "TimeUnix", "(ServiceName, MetricName, Attributes, toUnixTimestamp64Nano(TimeUnix))", "rand()"}

			if name == "otel_logs" {
				t.timestamp, t.orderBy = "Timestamp", "(toStartOfFiveMinutes(Timestamp), ServiceName, Timestamp)"
			} else if name == "otel_traces" {
				t.timestamp, t.orderBy, t.shardKey = "Timestamp", "(ServiceName, SpanName, toDateTime(Timestamp))", "cityHash64(TraceId)"
			}

			tables = append(tables, t)
		}
	}
	return tables, nil
}

// Prepare checks names on every replica and optionally creates missing objects.
// Cluster tables without the _local suffix must use the Distributed engine.
func (c *Client) Prepare(ctx context.Context, s Schema, createMissing bool) error {
	tables, err := s.tables()
	if err != nil {
		return err
	}

	nodes := int64(1)
	if s.Cluster != "" {
		nodes, err = c.count(ctx, "SELECT count() FROM "+s.systemTable("one"))
		if err != nil {
			return fmt.Errorf("check cluster: %w", err)
		}

		if nodes == 0 {
			return errors.New("cluster has no reachable replicas")
		}

		// Check engines first to avoid creating unused _local tables.
		if err := c.checkDistributedTables(ctx, s, tables); err != nil {
			return err
		}
	}

	databaseQuery := "SELECT count() FROM " + s.systemTable("databases") + " WHERE name = " + literal(s.Database)
	if err := c.ensure(ctx, "database "+s.Database, databaseQuery, nodes,
		"CREATE DATABASE IF NOT EXISTS "+identifier(s.Database)+s.onCluster(), createMissing, s.Cluster != ""); err != nil {
		return err
	}

	for _, t := range tables {
		local := t.name
		if s.Cluster != "" {
			local += "_local"
		}

		if err := c.ensureTable(ctx, s, local, nodes, s.localDDL(t, local), createMissing); err != nil {
			return err
		}

		if s.Cluster != "" {
			ddl := "CREATE TABLE IF NOT EXISTS " + s.qualified(t.name) + s.onCluster() +
				" AS " + s.qualified(local) + " ENGINE = Distributed(" + literal(s.Cluster) + ", " +
				literal(s.Database) + ", " + literal(local) + ", " + t.shardKey + ")"
			if err := c.ensureTable(ctx, s, t.name, nodes, ddl, createMissing); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Client) checkDistributedTables(ctx context.Context, s Schema, tables []table) error {
	names := make([]string, 0, len(tables))
	for _, t := range tables {
		names = append(names, literal(t.name))
	}

	query := "SELECT name, engine FROM " + s.systemTable("tables") +
		" WHERE database = " + literal(s.Database) + " AND name IN (" + strings.Join(names, ", ") +
		") AND engine != 'Distributed' LIMIT 1 FORMAT TabSeparated"

	data, err := c.query(ctx, query)
	if err != nil {
		return fmt.Errorf("check Distributed tables: %w", err)
	}

	if row := strings.TrimSpace(string(data)); row != "" {
		fields := strings.Split(row, "\t")
		if len(fields) != 2 {
			return errors.New("invalid ClickHouse table engine response")
		}

		return fmt.Errorf("cluster %q: table %s uses engine %s, expected Distributed; existing tables are not converted",
			s.Cluster, s.qualified(fields[0]), fields[1])
	}

	return nil
}

func (c *Client) ensureTable(ctx context.Context, s Schema, name string, nodes int64, ddl string, create bool) error {
	query := "SELECT count() FROM " + s.systemTable("tables") + " WHERE database = " + literal(s.Database) + " AND name = " + literal(name)

	return c.ensure(ctx, "table "+s.Database+"."+name, query, nodes, ddl, create, s.Cluster != "")
}

func (c *Client) ensure(ctx context.Context, object, query string, nodes int64, ddl string, create, clusterDDL bool) error {
	count, err := c.count(ctx, query)
	if err != nil {
		return fmt.Errorf("check %s: %w", object, err)
	}

	if count == nodes {
		return nil
	}

	if !create {
		return fmt.Errorf("%s is missing on one or more nodes (creation is disabled)", object)
	}

	if _, err := c.request(ctx, ddl, clusterDDL); err != nil {
		return fmt.Errorf("create %s: %w", object, err)
	}

	count, err = c.count(ctx, query)
	if err != nil {
		return fmt.Errorf("check created %s: %w", object, err)
	}

	if count != nodes {
		return fmt.Errorf("%s is still missing on one or more nodes after creation", object)
	}

	return nil
}

func (c *Client) count(ctx context.Context, query string) (int64, error) {
	data, err := c.query(ctx, query+" FORMAT TabSeparated")
	if err != nil {
		return 0, err
	}

	value, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return 0, errors.New("invalid ClickHouse count response")
	}

	return value, nil
}

func (s Schema) systemTable(name string) string {
	if s.Cluster != "" {
		return "clusterAllReplicas(" + literal(s.Cluster) + ", system." + name + ")"
	}

	return "system." + name
}

func (s Schema) qualified(name string) string { return identifier(s.Database) + "." + identifier(name) }
func (s Schema) onCluster() string {
	if s.Cluster == "" {
		return ""
	}

	return " ON CLUSTER " + literal(s.Cluster)
}

func (s Schema) localDDL(t table, name string) string {
	engine := "MergeTree()"
	if s.Replicated {
		engine = "ReplicatedMergeTree()"
	}

	ddl := "CREATE TABLE IF NOT EXISTS " + s.qualified(name) + s.onCluster() + " (\n" + t.columns +
		"\n) ENGINE = " + engine + " PARTITION BY " + s.Partition + "(" + t.timestamp + ") ORDER BY " + t.orderBy

	if s.TTL != 0 {
		clock := t.timestamp
		if s.Dataset == "otel" {
			clock = "toDateTime(" + clock + ")"
		}

		ddl += " TTL " + clock + " + toIntervalSecond(" + strconv.FormatInt(s.TTL, 10) + ")"
	}

	if s.Dataset == "otel" {
		ddl += " SETTINGS index_granularity = 8192, ttl_only_drop_parts = 1"
	}

	return ddl
}

func identifier(s string) string {
	return "`" + strings.NewReplacer(`\`, `\\`, "`", "\\`").Replace(s) + "`"
}

func literal(s string) string {
	return "'" + strings.NewReplacer(`\`, `\\`, "'", `\'`).Replace(s) + "'"
}
