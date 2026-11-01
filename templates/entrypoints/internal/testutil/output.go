package testutil

import (
	"os"
	"testing"
)

// Output capture changes process-wide streams; callers must not run in parallel.
func CaptureStdout(t *testing.T, run func()) string {
	t.Helper()
	return captureOutput(t, &os.Stdout, run)
}

func CaptureStderr(t *testing.T, run func()) string {
	t.Helper()
	return captureOutput(t, &os.Stderr, run)
}

func captureOutput(t *testing.T, stream **os.File, run func()) string {
	t.Helper()
	output, err := os.CreateTemp(t.TempDir(), "output")
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	original := *stream
	*stream = output
	defer func() { *stream = original }()
	run()
	data, err := os.ReadFile(output.Name())
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
