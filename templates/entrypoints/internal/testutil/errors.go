package testutil

import (
	"strings"
	"testing"
)

// CheckError fails the test unless err contains want; an empty want expects no error.
func CheckError(t *testing.T, err error, want string) {
	t.Helper()
	if want == "" {
		if err != nil {
			t.Fatal(err)
		}
		return
	}
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("error = %v, want %q", err, want)
	}
}
