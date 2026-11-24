package storageinit

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"testing/synctest"
	"time"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/testutil"
)

func TestPreparationRetries(t *testing.T) {
	refused := fmt.Errorf("check /dbl: %w", &net.OpError{Op: "dial", Net: "tcp", Err: syscall.ECONNREFUSED})
	for _, test := range []struct {
		name     string
		cause    error
		failures int
		calls    int
	}{
		{"ready on third attempt", refused, 2, 3},
		{"connection remains unavailable", refused, 3, 3},
		{"HTTP error", errors.New("Elasticsearch HTTP 403: forbidden"), 3, 1},
		{"SQL error", errors.New("ClickHouse HTTP 500: syntax error"), 3, 1},
		{"read error", &net.OpError{Op: "read", Net: "tcp", Err: syscall.ECONNRESET}, 3, 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				checkPreparationRetry(t, test.cause, test.failures, test.calls)
			})
		})
	}
}

// checkPreparationRetry runs a task that fails with cause the given number of times.
func checkPreparationRetry(t *testing.T, cause error, failures, wantCalls int) {
	t.Helper()
	calls := 0
	task := preparation{name: "ZBX_HISTORYPROVIDER_0", run: func(context.Context) error {
		calls++
		if calls <= failures {
			return cause
		}
		return nil
	}}
	start := time.Now()
	var err error
	logs := testutil.CaptureStderr(t, func() { err = task.prepare(context.Background()) })
	if calls != wantCalls {
		t.Fatalf("attempts = %d, want %d", calls, wantCalls)
	}
	if calls <= failures && !errors.Is(err, cause) || calls > failures && err != nil {
		t.Fatalf("unexpected result: %v", err)
	}
	if elapsed := time.Since(start); elapsed != time.Duration(calls-1)*10*time.Second {
		t.Fatalf("retry delay = %s", elapsed)
	}
	if strings.Count(logs, "retrying in 10s") != calls-1 || calls > 1 && !strings.Contains(logs, task.name) {
		t.Fatalf("unexpected retry log: %s", logs)
	}
}

func TestCancelRetryWait(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		calls := 0
		task := preparation{run: func(context.Context) error {
			calls++
			return &net.OpError{Op: "dial", Err: syscall.ECONNREFUSED}
		}}
		go func() {
			time.Sleep(time.Second)
			cancel()
		}()
		started := time.Now()
		if err := task.prepare(ctx); !errors.Is(err, context.Canceled) {
			t.Fatalf("expected cancellation, got %v", err)
		}
		if calls != 1 || time.Since(started) != time.Second {
			t.Fatal("cancellation did not stop the retry wait")
		}
	})
}

func TestRestartAfterPartialCreation(t *testing.T) {
	var state partialCreation
	state.failUint.Store(true)
	ch := httptest.NewServer(state.clickHouse(t))
	defer ch.Close()
	es := httptest.NewServer(state.elasticsearch(t))
	defer es.Close()
	env := bootstrap.Environment{
		"ZBX_HISTORYPROVIDER_0":        fmt.Sprintf(`{"provider":"clickhouse","url":%q,"db":"history","types":["str"]}`, ch.URL),
		"ZBX_HISTORYPROVIDER_0_SCHEMA": `{"create":true}`,
		"ZBX_HISTORYPROVIDER_1":        fmt.Sprintf(`{"provider":"elasticsearch","url":%q,"types":["dbl","uint"]}`, es.URL),
		"ZBX_HISTORYPROVIDER_1_SCHEMA": `{"create":true}`,
	}
	testutil.CheckError(t, ConfigureHistory(maps.Clone(env)), "ZBX_HISTORYPROVIDER_1")
	if !state.history.Load() || !state.dbl.Load() || state.uintTable.Load() {
		t.Fatal("expected ClickHouse and the first Elasticsearch index to be created")
	}
	state.failUint.Store(false)
	if err := ConfigureHistory(maps.Clone(env)); err != nil {
		t.Fatal(err)
	}
	if !state.uintTable.Load() || state.historyCreates.Load() != 1 || state.dblCreates.Load() != 1 {
		t.Fatal("restart must create the missing index without recreating existing tables or indices")
	}
}

// partialCreation tracks objects created by fake ClickHouse and Elasticsearch servers.
type partialCreation struct {
	database, history, dbl, uintTable, failUint atomic.Bool
	historyCreates, dblCreates                  atomic.Int32
}

func (s *partialCreation) clickHouse(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		query := string(body)
		switch {
		case strings.HasPrefix(query, "CREATE DATABASE"):
			s.database.Store(true)
		case strings.HasPrefix(query, "CREATE TABLE"):
			s.history.Store(true)
			s.historyCreates.Add(1)
		case strings.HasPrefix(query, "SELECT count()"):
			exists := s.history.Load()
			if strings.Contains(query, "system.databases") {
				exists = s.database.Load()
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
	}
}

func (s *partialCreation) elasticsearch(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var exists *atomic.Bool
		switch r.URL.Path {
		case "/dbl":
			exists = &s.dbl
		case "/uint":
			exists = &s.uintTable
		default:
			t.Errorf("unexpected Elasticsearch path: %s", r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if r.Method == http.MethodPut {
			if r.URL.Path == "/uint" && s.failUint.Load() {
				http.Error(w, "creation failed", http.StatusServiceUnavailable)
				return
			}
			if r.URL.Path == "/dbl" {
				s.dblCreates.Add(1)
			}
			exists.Store(true)
		} else if !exists.Load() {
			w.WriteHeader(http.StatusNotFound)
		}
	}
}
