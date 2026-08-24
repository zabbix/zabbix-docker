package bootstrap

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestUpdateConfigValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), "agent.conf")
	if err := os.WriteFile(path, []byte("# DenyKey=system.run[*]\nOther=value\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := UpdateConfigValues(path, "DenyKey", `"one,two"`); err != nil {
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

func TestUpdateConfigValuesNormalizesLineEndings(t *testing.T) {
	path := filepath.Join(t.TempDir(), "agent.conf")
	if err := os.WriteFile(path, []byte("# DenyKey=system.run[*]\r\nOther=value\r\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := UpdateConfigValues(path, "DenyKey", "one,two"); err != nil {
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

func TestUpdateConfigValuesPreservesActiveValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), "agent.conf")
	if err := os.WriteFile(path, []byte("DenyKey=existing\n# DenyKey=system.run[*]\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := UpdateConfigValues(path, "DenyKey", "existing,new"); err != nil {
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

func TestUpdateConfigValuesRemovesActiveValuesWhenEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "agent.conf")
	if err := os.WriteFile(path, []byte("DenyKey=system.run[*]\nOther=value\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := UpdateConfigValues(path, "DenyKey", ""); err != nil {
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

func TestUpdateConfigValuesDoesNotRewriteUnchangedConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "agent.conf")
	data := []byte("# DenyKey=system.run[*]\nOther=value\n")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	modified := time.Unix(1_700_000_000, 0)
	if err := os.Chtimes(path, modified, modified); err != nil {
		t.Fatal(err)
	}

	if err := UpdateConfigValues(path, "DenyKey", ""); err != nil {
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
	for _, value := range []string{
		"HistoryProvider=${ZBX_HISTORYPROVIDER_0}",
		"HistoryProvider=${ZBX_HISTORYPROVIDER_1}",
	} {
		if !strings.Contains(string(data), value) {
			t.Fatalf("configuration does not contain %q:\n%s", value, data)
		}
	}
	if env["ZBX_HISTORYPROVIDER_0"] == "" || env["ZBX_HISTORYPROVIDER_1"] == "" {
		t.Fatal("referenced indexed variables were removed")
	}
	if env["ZBX_HISTORYPROVIDER_3"] != "ignored-after-gap" {
		t.Fatal("indexed variable after a gap was unexpectedly processed")
	}
}

func TestUpdateConfigIndexedLogging(t *testing.T) {
	path := filepath.Join(t.TempDir(), "agent.conf")
	if err := os.WriteFile(path, []byte("# UserParameter=\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	env := Environment{
		"ZBX_USERPARAMETER_0": `custom.token,printf '%s' 'sensitive-token'`,
		"ZBX_USERPARAMETER_1": `custom.password,printf '%s' 'sensitive-password'`,
	}

	var updateErr error
	output := captureStdout(t, func() {
		updateErr = UpdateConfigIndexed(env, path, "UserParameter", "ZBX_USERPARAMETER")
	})
	if updateErr != nil {
		t.Fatal(updateErr)
	}
	if output != "" {
		t.Fatalf("indexed update logged at info level: %s", output)
	}

	env["DEBUG_MODE"] = "true"
	output = captureStdout(t, func() {
		updateErr = UpdateConfigIndexed(env, path, "UserParameter", "ZBX_USERPARAMETER")
	})
	if updateErr != nil {
		t.Fatal(updateErr)
	}
	for _, expected := range []string{
		"[debug]",
		"2 indexed environment variables",
		"ZBX_USERPARAMETER_0, ZBX_USERPARAMETER_1",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("debug output does not contain %q: %s", expected, output)
		}
	}
	for _, sensitive := range []string{"sensitive-token", "sensitive-password"} {
		if strings.Contains(output, sensitive) {
			t.Fatalf("debug output contains sensitive value %q: %s", sensitive, output)
		}
	}
}

func captureStdout(t *testing.T, run func()) string {
	t.Helper()

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	stdout := os.Stdout
	os.Stdout = writer
	defer func() {
		os.Stdout = stdout
	}()

	run()
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}

	return string(output)
}
