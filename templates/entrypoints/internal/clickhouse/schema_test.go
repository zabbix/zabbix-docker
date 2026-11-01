package clickhouse

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	c, err := NewClient(Connection{URL: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(c.Close)
	return c
}

func TestExistingTables(t *testing.T) {
	for _, create := range []bool{false, true} {
		t.Run(fmt.Sprint(create), func(t *testing.T) {
			var queries []string
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				queries = append(queries, string(body))
				fmt.Fprintln(w, "1")
			})
			err := c.Prepare(context.Background(), Schema{Database: "existing", Dataset: "history", ValueTypes: []string{"uint", "log"}, Partition: "toDate"}, create)
			if err != nil {
				t.Fatal(err)
			}
			if len(queries) != 3 {
				t.Fatalf("queries = %v", queries)
			}
			for _, q := range queries {
				if !strings.HasPrefix(q, "SELECT count() FROM system.") || strings.Contains(q, "columns") || strings.Contains(q, "engine") {
					t.Fatalf("unexpected structure inspection or mutation: %s", q)
				}
			}
		})
	}
}

func TestCreationDisabled(t *testing.T) {
	calls := 0
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		fmt.Fprintln(w, "0")
	})
	err := c.Prepare(context.Background(), Schema{Database: "missing", Dataset: "otel", Partition: "toDate"}, false)
	if err == nil || !strings.Contains(err.Error(), "creation is disabled") || calls != 1 {
		t.Fatalf("error=%v, calls=%d", err, calls)
	}
}

func TestEnsureChecksDDLCompletion(t *testing.T) {
	for _, replicas := range []int{1, 2} {
		t.Run(fmt.Sprintf("replicas=%d", replicas), func(t *testing.T) {
			created := false
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				if strings.HasPrefix(string(body), "CREATE") {
					created = true
					return
				}
				count := 1
				if created {
					count = replicas
				}
				fmt.Fprintln(w, count)
			})
			err := c.ensure(context.Background(), "database data", "SELECT count() FROM system.databases", 2,
				"CREATE DATABASE IF NOT EXISTS data ON CLUSTER 'cluster'", true, true)
			if !created {
				t.Fatal("missing database was not created")
			}
			if replicas == 2 && err != nil {
				t.Fatal(err)
			}
			if replicas == 1 && (err == nil || !strings.Contains(err.Error(), "still missing")) {
				t.Fatalf("expected an error for incomplete DDL, got %v", err)
			}
		})
	}
}

func TestLocalDDLUsesSchemaOptions(t *testing.T) {
	s := Schema{Database: "data", Dataset: "history", ValueTypes: []string{"uint"}, Cluster: "cluster", Replicated: true, TTL: 3600, Partition: "toYYYYMM"}
	tables, err := s.tables()
	if err != nil {
		t.Fatal(err)
	}
	ddl := s.localDDL(tables[0], "history_uint_local")
	for _, fragment := range []string{"IF NOT EXISTS", "ON CLUSTER 'cluster'", "ReplicatedMergeTree()", "toYYYYMM(clock_ns)", "toIntervalSecond(3600)"} {
		if !strings.Contains(ddl, fragment) {
			t.Errorf("DDL does not contain %q: %s", fragment, ddl)
		}
	}
}

func TestOTelIncludesSummaryAndShardKeys(t *testing.T) {
	s := Schema{Database: "otel", Dataset: "otel", Partition: "toDate", TTL: 2678400}
	tables, err := s.tables()
	if err != nil {
		t.Fatal(err)
	}
	if len(tables) != 7 {
		t.Fatalf("table count = %d", len(tables))
	}
	for _, table := range tables {
		if table.name == "otel_traces" && table.shardKey != "cityHash64(TraceId)" {
			t.Fatal("trace spans must use the same shard key")
		}
		if table.name != "otel_traces" && table.shardKey != "rand()" {
			t.Fatal("unexpected sharding key")
		}
		if table.name == "otel_metrics_summary" && !strings.Contains(table.columns, `"ValueAtQuantiles.Quantile" Array(Float64)`) {
			t.Fatal("missing Summary columns")
		}
	}
}

func TestSQLNamesAreQuoted(t *testing.T) {
	if got := literal("a'\\b"); got != `'a\'\\b'` {
		t.Fatal(got)
	}
	if got := identifier("a`\\b"); got != "`a\\`\\\\b`" {
		t.Fatal(got)
	}
}

func TestClusterRejectsNonDistributedTableBeforeDDL(t *testing.T) {
	for _, dataset := range []string{"history", "otel"} {
		for _, create := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/create=%t", dataset, create), func(t *testing.T) {
				s := Schema{Database: "data", Dataset: dataset, ValueTypes: []string{"uint", "log"}, Cluster: "cluster", Partition: "toDate"}
				name := "otel_metrics_summary"
				if dataset == "history" {
					name = "history_log"
				}
				calls := 0
				c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
					calls++
					body, _ := io.ReadAll(r.Body)
					q := string(body)
					switch {
					case strings.Contains(q, "system.one"):
						fmt.Fprintln(w, "4")
					case strings.HasPrefix(q, "SELECT name, engine"):
						if !strings.Contains(q, "clusterAllReplicas('cluster', system.tables)") ||
							!strings.Contains(q, literal(name)) || !strings.Contains(q, "engine != 'Distributed'") {
							t.Errorf("engine check must cover the requested tables on all replicas: %s", q)
						}
						fmt.Fprintf(w, "%s\tMergeTree\n", name)
					default:
						t.Errorf("unexpected query while checking table engines: %s", q)
					}
				})
				err := c.Prepare(context.Background(), s, create)
				if err == nil || !strings.Contains(err.Error(), name) || !strings.Contains(err.Error(), "expected Distributed") || calls != 2 {
					t.Fatalf("calls=%d error=%v", calls, err)
				}
			})
		}
	}
}

func TestExistingDistributedTables(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		q := string(body)
		if strings.HasPrefix(q, "SELECT name, engine") {
			return
		}
		if !strings.HasPrefix(q, "SELECT count()") {
			t.Errorf("unexpected DDL: %s", q)
		}
		fmt.Fprintln(w, "4")
	})
	for _, create := range []bool{false, true} {
		if err := c.Prepare(context.Background(), Schema{Database: "data", Dataset: "history", ValueTypes: []string{"uint"}, Cluster: "cluster", Partition: "toDate"}, create); err != nil {
			t.Fatal(err)
		}
	}
}
