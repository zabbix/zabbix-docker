package bootstrap

import (
	"context"
	"errors"
	"time"
)

// RetryOptions describes how Retry repeats a failing attempt.
type RetryOptions struct {
	// Attempts limits the number of calls. Zero repeats the attempt until it
	// succeeds or the context ends, which is what the entrypoints do while they
	// wait for a service to start.
	Attempts int
	// Interval is the pause between attempts.
	Interval time.Duration
	// OnRetry reports the failure that is about to be retried, so that the
	// caller can log it in its own words.
	OnRetry func(error)
	// Wait replaces the pause between attempts. Tests use it to observe the
	// delay instead of sleeping.
	Wait func(context.Context, time.Duration) error
}

// Retry calls attempt until it succeeds, pausing Interval between calls. It
// stops early when the context ends or when a failure is marked with Stop, and
// returns the last failure once the attempts run out.
func Retry(ctx context.Context, options RetryOptions, attempt func() error) error {
	wait := options.Wait
	if wait == nil {
		wait = Wait
	}

	var lastErr error
	for call := 0; options.Attempts == 0 || call < options.Attempts; call++ {
		if call > 0 {
			if options.OnRetry != nil {
				options.OnRetry(lastErr)
			}
			if err := wait(ctx, options.Interval); err != nil {
				return err
			}
		}

		lastErr = attempt()
		if lastErr == nil {
			return nil
		}

		var stop stopError
		if errors.As(lastErr, &stop) {
			return stop.error
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	return lastErr
}

// Stop marks a failure that Retry must not repeat. The message of err is kept,
// and errors.Is and errors.As still match what it wraps.
func Stop(err error) error {
	return stopError{err}
}

type stopError struct {
	error
}

func (s stopError) Unwrap() error {
	return s.error
}

// Wait pauses for interval unless the context ends first.
func Wait(ctx context.Context, interval time.Duration) error {
	if interval <= 0 {
		return ctx.Err()
	}

	timer := time.NewTimer(interval)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
