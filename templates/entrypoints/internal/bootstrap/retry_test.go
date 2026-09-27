package bootstrap

import (
	"context"
	"errors"
	"testing"
	"time"
)

func noWait() func(context.Context, time.Duration) error {
	return func(context.Context, time.Duration) error { return nil }
}

func TestRetryRepeatsUntilSuccess(t *testing.T) {
	calls := 0
	retries := 0
	err := Retry(context.Background(), RetryOptions{
		Interval: time.Hour,
		OnRetry:  func(error) { retries++ },
		Wait:     noWait(),
	}, func() error {
		calls++
		if calls < 3 {
			return errors.New("not ready")
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 3 || retries != 2 {
		t.Fatalf("calls = %d, retries = %d, want 3 and 2", calls, retries)
	}
}

func TestRetryGivesUpAfterAttempts(t *testing.T) {
	calls := 0
	last := errors.New("connection refused")
	err := Retry(context.Background(), RetryOptions{Attempts: 3, Wait: noWait()}, func() error {
		calls++

		return last
	})
	if !errors.Is(err, last) {
		t.Fatalf("error = %v, want %v", err, last)
	}
	if calls != 3 {
		t.Fatalf("calls = %d, want 3", calls)
	}
}

func TestRetryKeepsTheMessageOfAStoppedFailure(t *testing.T) {
	calls := 0
	fatal := errors.New("access denied")
	err := Retry(context.Background(), RetryOptions{Wait: noWait()}, func() error {
		calls++

		return Stop(fatal)
	})
	if !errors.Is(err, fatal) || err.Error() != fatal.Error() {
		t.Fatalf("error = %v, want %v", err, fatal)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
}

func TestRetryStopsWithTheContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	err := Retry(ctx, RetryOptions{Interval: time.Hour}, func() error {
		calls++
		cancel()

		return errors.New("not ready")
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want %v", err, context.Canceled)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
}

func TestWaitReturnsWhenTheContextEnds(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := Wait(ctx, time.Hour); !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want %v", err, context.Canceled)
	}
	if err := Wait(context.Background(), time.Millisecond); err != nil {
		t.Fatal(err)
	}
}
