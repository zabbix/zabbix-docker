package bootstrap

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// TerminationContext returns a context canceled by an interrupt or
// termination signal.
func TerminationContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
}
