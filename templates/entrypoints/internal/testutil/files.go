package testutil

import (
	"compress/gzip"
	"os"
	"testing"
)

// WriteGzipFile writes a compressed test fixture.
func WriteGzipFile(t *testing.T, path, content string) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	compressed := gzip.NewWriter(file)
	if _, err := compressed.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := compressed.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}
