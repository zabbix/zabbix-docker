package storageinit

import (
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

func TestMixedHistorySchemas(t *testing.T) {
	chCalls, esCalls := 0, 0
	ch := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { chCalls++; fmt.Fprintln(w, "1") }))
	defer ch.Close()
	es := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		esCalls++
		if r.Method != "HEAD" || r.URL.Path != "/str" {
			t.Errorf("unexpected Elasticsearch request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer es.Close()
	env := bootstrap.Environment{
		"ZBX_HISTORYPROVIDER_0":          fmt.Sprintf(`{"provider":"clickhouse","url":%q,"db":"history","types":["uint"]}`, ch.URL),
		"ZBX_HISTORYPROVIDER_0_SCHEMA":   `{"create":false}`,
		"ZBX_HISTORYPROVIDER_1":          fmt.Sprintf(`{"provider":"elasticsearch","url":%q,"types":["str"]}`, es.URL),
		"ZBX_HISTORYPROVIDER_1_SCHEMA":   `{"create":false}`,
		"ZBX_TELEMETRYPROVIDER_0_SCHEMA": "invalid telemetry _SCHEMA",
	}
	if err := ConfigureHistory(env); err != nil {
		t.Fatal(err)
	}
	if chCalls == 0 || esCalls == 0 {
		t.Fatalf("history providers were not both checked: ClickHouse=%d Elasticsearch=%d", chCalls, esCalls)
	}
	for name, provider := range map[string]string{
		"ZBX_HISTORYPROVIDER_0": "clickhouse",
		"ZBX_HISTORYPROVIDER_1": "elasticsearch",
	} {
		if !strings.HasPrefix(env[name], provider+";") {
			t.Errorf("%s was not converted to native configuration: %q", name, env[name])
		}
		if _, ok := env[name+"_SCHEMA"]; ok {
			t.Errorf("%s_SCHEMA was not removed", name)
		}
	}
}

func TestElasticsearchOptions(t *testing.T) {
	for _, raw := range []string{`{}`, `null`, `{"create":true,"cluster":"x"}`, `{"create":true,"shards":0}`, `{"create":true,"replicas":-1}`, `{"create":false} {}`} {
		t.Run(raw, func(t *testing.T) {
			env := bootstrap.Environment{"ZBX_HISTORYPROVIDER_0": `{"provider":"elasticsearch","url":"http://unused:9200","types":["str"]}`, "ZBX_HISTORYPROVIDER_0_SCHEMA": raw}
			prepareCADirs(t, env)
			if _, err := PrepareHistory(env); err == nil {
				t.Fatal("invalid schema accepted")
			}
		})
	}
	for _, value := range []string{`"daily"`, `2`} {
		env := bootstrap.Environment{"ZBX_HISTORYPROVIDER_0": `{"provider":"elasticsearch","url":"http://unused:9200","types":["str"],"date_index":` + value + `}`, "ZBX_HISTORYPROVIDER_0_SCHEMA": `{"create":true}`}
		prepareCADirs(t, env)
		if _, err := PrepareHistory(env); err == nil {
			t.Fatal("invalid date_index accepted")
		}
	}
	env := bootstrap.Environment{"ZBX_HISTORYPROVIDER_0": `{"provider":"elasticsearch","url":"http://unused:9200","types":["str"]}`}
	if count, err := PrepareHistory(env); err != nil || count != 0 {
		t.Fatalf("provider without _SCHEMA: count=%d error=%v", count, err)
	}
}

func TestRestartAfterPartialCreation(t *testing.T) {
	var database, history, dbl, uintTable atomic.Bool
	var historyCreates, dblCreates atomic.Int32
	var failUint atomic.Bool
	failUint.Store(true)
	ch := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		query := string(body)
		switch {
		case strings.HasPrefix(query, "CREATE DATABASE"):
			database.Store(true)
		case strings.HasPrefix(query, "CREATE TABLE"):
			history.Store(true)
			historyCreates.Add(1)
		case strings.HasPrefix(query, "SELECT count()"):
			exists := history.Load()
			if strings.Contains(query, "system.databases") {
				exists = database.Load()
			}
			if exists {
				fmt.Fprintln(w, "1")
			} else {
				fmt.Fprintln(w, "0")
			}
		default:
			t.Errorf("unexpected ClickHouse query: %s", query)
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer ch.Close()
	es := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var exists *atomic.Bool
		switch r.URL.Path {
		case "/dbl":
			exists = &dbl
		case "/uint":
			exists = &uintTable
		default:
			t.Errorf("unexpected Elasticsearch path: %s", r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if r.Method == http.MethodPut {
			if r.URL.Path == "/uint" && failUint.Load() {
				http.Error(w, "creation failed", http.StatusServiceUnavailable)
				return
			}
			if r.URL.Path == "/dbl" {
				dblCreates.Add(1)
			}
			exists.Store(true)
		} else if !exists.Load() {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer es.Close()
	env := bootstrap.Environment{
		"ZBX_HISTORYPROVIDER_0":        fmt.Sprintf(`{"provider":"clickhouse","url":%q,"db":"history","types":["str"]}`, ch.URL),
		"ZBX_HISTORYPROVIDER_0_SCHEMA": `{"create":true}`,
		"ZBX_HISTORYPROVIDER_1":        fmt.Sprintf(`{"provider":"elasticsearch","url":%q,"types":["dbl","uint"]}`, es.URL),
		"ZBX_HISTORYPROVIDER_1_SCHEMA": `{"create":true}`,
	}
	if err := ConfigureHistory(maps.Clone(env)); err == nil || !strings.Contains(err.Error(), "ZBX_HISTORYPROVIDER_1") {
		t.Fatalf("expected second provider failure, got %v", err)
	}
	if !history.Load() || !dbl.Load() || uintTable.Load() {
		t.Fatal("expected ClickHouse and the first Elasticsearch index to be created")
	}
	failUint.Store(false)
	if err := ConfigureHistory(maps.Clone(env)); err != nil {
		t.Fatal(err)
	}
	if !uintTable.Load() || historyCreates.Load() != 1 || dblCreates.Load() != 1 {
		t.Fatal("restart must create the missing index without recreating existing tables or indices")
	}
}
