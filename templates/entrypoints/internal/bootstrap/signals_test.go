//go:build !windows

package bootstrap

import (
	"context"
	"errors"
	"os"
	"testing"
)

func TestTerminationContext(t *testing.T) {
	ctx, stop := TerminationContext()
	defer stop()

	process, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	if err := process.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}
	<-ctx.Done()
	if !errors.Is(ctx.Err(), context.Canceled) {
		t.Fatalf("TerminationContext() error = %v, want context.Canceled", ctx.Err())
	}
}
