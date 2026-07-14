package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestUpdateConfigMultiple(t *testing.T) {
	path := filepath.Join(t.TempDir(), "agent.conf")
	if err := os.WriteFile(path, []byte("# DenyKey=system.run[*]\nOther=value\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := UpdateConfigMultiple(path, "DenyKey", `"one,two"`); err != nil {
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
}

func TestUpdateConfigMultiplePreservesActiveValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), "agent.conf")
	if err := os.WriteFile(path, []byte("DenyKey=existing\n# DenyKey=system.run[*]\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := UpdateConfigMultiple(path, "DenyKey", "existing,new"); err != nil {
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

func TestUpdateConfigMultipleRemovesActiveValuesWhenEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "agent.conf")
	if err := os.WriteFile(path, []byte("DenyKey=system.run[*]\nOther=value\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := UpdateConfigMultiple(path, "DenyKey", ""); err != nil {
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

func TestUpdateConfigMultipleDoesNotRewriteUnchangedConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "agent.conf")
	data := []byte("# DenyKey=system.run[*]\nOther=value\n")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	modified := time.Unix(1_700_000_000, 0)
	if err := os.Chtimes(path, modified, modified); err != nil {
		t.Fatal(err)
	}

	if err := UpdateConfigMultiple(path, "DenyKey", ""); err != nil {
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

func TestUpdateConfigIndexed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "server.conf")
	if err := os.WriteFile(path, []byte("# HistoryProvider=\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	env := Environment{
		"ZBX_HISTORYPROVIDER_0": "provider-one",
		"ZBX_HISTORYPROVIDER_1": "provider-two",
		"ZBX_HISTORYPROVIDER_3": "ignored-after-gap",
	}
	if err := UpdateConfigIndexed(env, path, "HistoryProvider", "ZBX_HISTORYPROVIDER"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"HistoryProvider=provider-one", "HistoryProvider=provider-two"} {
		if !strings.Contains(string(data), value) {
			t.Fatalf("configuration does not contain %q:\n%s", value, data)
		}
	}
	if _, found := env["ZBX_HISTORYPROVIDER_0"]; found {
		t.Fatal("processed indexed variable was not removed")
	}
	if env["ZBX_HISTORYPROVIDER_3"] != "ignored-after-gap" {
		t.Fatal("indexed variable after a gap was unexpectedly processed")
	}
}
