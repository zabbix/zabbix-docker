//go:build windows

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMergeParameterValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), "agent.conf")
	if err := os.WriteFile(path, []byte("# DenyKey=system.run[*]\nOther=value\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := MergeParameterValues(path, "DenyKey", `"one,two"`); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := "# DenyKey=system.run[*]\nDenyKey=one\nDenyKey=two\nOther=value\n"
	if string(data) != want {
		t.Fatalf("config:\n%s\nwant:\n%s", data, want)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("config mode = %o, want 600", info.Mode().Perm())
	}
}

func TestMergeParameterValuesNormalizesCRLF(t *testing.T) {
	path := filepath.Join(t.TempDir(), "agent.conf")
	if err := os.WriteFile(path, []byte("# DenyKey=system.run[*]\r\nOther=value\r\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := MergeParameterValues(path, "DenyKey", "one,two"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := "# DenyKey=system.run[*]\nDenyKey=one\nDenyKey=two\nOther=value\n"
	if string(data) != want {
		t.Fatalf("config:\n%q\nwant:\n%q", data, want)
	}
}

func TestMergeParameterValuesPreservesActiveValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), "agent.conf")
	if err := os.WriteFile(path, []byte("DenyKey=existing\n# DenyKey=system.run[*]\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := MergeParameterValues(path, "DenyKey", "existing,new"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if strings.Count(content, "DenyKey=existing\n") != 1 {
		t.Fatalf("existing value was removed or duplicated:\n%s", data)
	}
	if !strings.Contains(content, "DenyKey=new\n") {
		t.Fatalf("new value was not added:\n%s", data)
	}
}

func TestMergeParameterValuesRemovesActiveValuesWhenEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "agent.conf")
	if err := os.WriteFile(path, []byte("DenyKey=system.run[*]\nOther=value\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := MergeParameterValues(path, "DenyKey", ""); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "DenyKey=") {
		t.Fatalf("active DenyKey was not removed:\n%s", data)
	}
}

func TestMergeParameterValuesDoesNotRewriteUnchangedConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "agent.conf")
	data := []byte("# DenyKey=system.run[*]\nOther=value\n")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	modified := time.Unix(1_700_000_000, 0)
	if err := os.Chtimes(path, modified, modified); err != nil {
		t.Fatal(err)
	}

	if err := MergeParameterValues(path, "DenyKey", ""); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !info.ModTime().Equal(modified) {
		t.Fatalf("unchanged configuration was rewritten: modification time is %s", info.ModTime())
	}
}

func TestSetParameterSetsUndefinedValue(t *testing.T) {
	path := filepath.Join(t.TempDir(), "agent.conf")
	if err := os.WriteFile(path, []byte("Hostname=configured\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := SetParameter(path, "Hostname", `""`); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "Hostname=\n" {
		t.Fatalf("config: %q, want an undefined Hostname", data)
	}
}
