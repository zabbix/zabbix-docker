package storageinit

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"syscall"
	"testing"
	"testing/synctest"
	"time"

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
				calls := 0
				task := preparation{name: "ZBX_HISTORYPROVIDER_0", run: func(context.Context) error {
					calls++
					if calls <= test.failures {
						return test.cause
					}
					return nil
				}}
				start := time.Now()
				var err error
				logs := testutil.CaptureStderr(t, func() { err = task.prepare(context.Background()) })
				if calls != test.calls {
					t.Fatalf("attempts = %d, want %d", calls, test.calls)
				}
				if calls <= test.failures && !errors.Is(err, test.cause) || calls > test.failures && err != nil {
					t.Fatalf("unexpected result: %v", err)
				}
				if elapsed := time.Since(start); elapsed != time.Duration(calls-1)*10*time.Second {
					t.Fatalf("retry delay = %s", elapsed)
				}
				if strings.Count(logs, "retrying in 10s") != calls-1 || calls > 1 && !strings.Contains(logs, task.name) {
					t.Fatalf("unexpected retry log: %s", logs)
				}
			})
		})
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
