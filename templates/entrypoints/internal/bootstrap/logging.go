package bootstrap

import (
	"fmt"
	"os"
	"time"
)

func LogInfo(format string, args ...any) {
	logMessage(os.Stdout, "info", format, args...)
}

func LogDebug(env Environment, format string, args ...any) {
	if env["DEBUG_MODE"] != "true" {
		return
	}

	logMessage(os.Stdout, "debug", format, args...)
}

func LogWarn(format string, args ...any) {
	logMessage(os.Stderr, "warning", format, args...)
}

func LogError(format string, args ...any) {
	logMessage(os.Stderr, "error", format, args...)
}

func logMessage(file *os.File, level, format string, args ...any) {
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	fmt.Fprintf(file, "%s [%s]: %s\n", timestamp, level, fmt.Sprintf(format, args...))
}
