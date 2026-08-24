package config

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

func TestConfigureServers(t *testing.T) {
	env := bootstrap.Environment{
		"ZBX_SERVER_HOST": "server", "ZBX_SERVER_PORT": "10061",
		"ZBX_PASSIVESERVERS": "passive", "ZBX_ACTIVESERVERS": "active",
	}
	ConfigureServers(env)
	if env["ZBX_PASSIVESERVERS"] != "server,passive" || env["ZBX_ACTIVESERVERS"] != "server:10061,active" {
		t.Fatalf("unexpected servers: %#v", env)
	}
}

func TestConfigureItemKeyRules(t *testing.T) {
	configDir := t.TempDir()
	path := filepath.Join(configDir, "item_keys.conf")
	initial := "header\n# AllowKey=\nAllowKey=old.allow\nmiddle\nDenyKeyRegexp=old.deny\nfooter\n"
	if err := os.WriteFile(path, []byte(initial), 0o600); err != nil {
		t.Fatal(err)
	}

	env := bootstrap.Environment{
		"DEBUG_MODE":            "true",
		"ZBX_ALLOWKEY_0":        "sensitive-allow",
		"ZBX_DENYKEY_1":         "sensitive-deny",
		"ZBX_ALLOWKEY_REGEXP_2": "sensitive-allow-regexp",
		"ZBX_DENYKEY_REGEXP_3":  "sensitive-deny-regexp",
	}

	var configureErr error
	logOutput := captureAgentStdout(t, func() {
		configureErr = ConfigureItemKeyRules(env, configDir, "item_keys.conf")
	})
	if configureErr != nil {
		t.Fatal(configureErr)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := "header\n# AllowKey=\nmiddle\nfooter\n\n" +
		"AllowKey=${ZBX_ALLOWKEY_0}\n" +
		"DenyKey=${ZBX_DENYKEY_1}\n" +
		"AllowKeyRegexp=${ZBX_ALLOWKEY_REGEXP_2}\n" +
		"DenyKeyRegexp=${ZBX_DENYKEY_REGEXP_3}\n"
	if string(data) != want {
		t.Fatalf("item key config:\n%s\nwant:\n%s", data, want)
	}

	for _, variable := range []string{
		"ZBX_ALLOWKEY_0", "ZBX_DENYKEY_1",
		"ZBX_ALLOWKEY_REGEXP_2", "ZBX_DENYKEY_REGEXP_3",
	} {
		if !strings.Contains(logOutput, variable) {
			t.Fatalf("debug output does not contain %s: %s", variable, logOutput)
		}
		if env[variable] == "" {
			t.Fatalf("%s was removed from the environment", variable)
		}
	}
	for _, sensitive := range []string{
		"sensitive-allow", "sensitive-deny",
		"sensitive-allow-regexp", "sensitive-deny-regexp",
	} {
		if strings.Contains(logOutput, sensitive) {
			t.Fatalf("debug output contains sensitive value %q: %s", sensitive, logOutput)
		}
	}
}

func TestConfigureItemKeyRulesNormalizesLineEndings(t *testing.T) {
	configDir := t.TempDir()
	path := filepath.Join(configDir, "item_keys.conf")
	if err := os.WriteFile(path, []byte("header\r\nAllowKey=old\r\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	env := bootstrap.Environment{"ZBX_ALLOWKEY_0": "system.localtime"}
	if err := ConfigureItemKeyRules(env, configDir, "item_keys.conf"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := "header\n\nAllowKey=${ZBX_ALLOWKEY_0}\n"
	if string(data) != want {
		t.Fatalf("item key config:\n%q\nwant:\n%q", data, want)
	}
}

func TestConfigureItemKeyRulesValidation(t *testing.T) {
	tests := []struct {
		name string
		env  bootstrap.Environment
		want string
	}{
		{
			name: "legacy variable",
			env:  bootstrap.Environment{"ZBX_ALLOWKEY": "system.localtime"},
			want: "ZBX_ALLOWKEY is not supported",
		},
		{
			name: "empty legacy variable",
			env:  bootstrap.Environment{"ZBX_ALLOWKEY": ""},
			want: "ZBX_ALLOWKEY is not supported",
		},
		{
			name: "regexp variable without index",
			env:  bootstrap.Environment{"ZBX_DENYKEY_REGEXP": "^system\\.run"},
			want: "ZBX_DENYKEY_REGEXP is not supported",
		},
		{
			name: "duplicate index",
			env: bootstrap.Environment{
				"ZBX_ALLOWKEY_0": "system.localtime",
				"ZBX_DENYKEY_0":  "system.run[*]",
			},
			want: "index 0 is used by both",
		},
		{
			name: "missing index",
			env: bootstrap.Environment{
				"ZBX_ALLOWKEY_0": "system.localtime",
				"ZBX_DENYKEY_2":  "system.run[*]",
			},
			want: "ZBX_DENYKEY_2 uses index 2, but index 1 is missing",
		},
		{
			name: "empty rule",
			env:  bootstrap.Environment{"ZBX_ALLOWKEY_0": ""},
			want: "ZBX_ALLOWKEY_0 must not be empty",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ConfigureItemKeyRules(test.env, t.TempDir(), "item_keys.conf")
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want containing %q", err, test.want)
			}
		})
	}
}

func TestConfigureItemKeyRulesWithoutVariablesRemovesExistingRules(t *testing.T) {
	configDir := t.TempDir()
	path := filepath.Join(configDir, "item_keys.conf")
	if err := os.WriteFile(path, []byte("AllowKey=custom.key\nOther=value\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := ConfigureItemKeyRules(bootstrap.Environment{}, configDir, "item_keys.conf"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "Other=value\n" {
		t.Fatalf("existing item key rules were not removed:\n%s", data)
	}
}

func TestConfigureItemKeyRulesWithoutVariablesRequiresConfigFile(t *testing.T) {
	err := ConfigureItemKeyRules(bootstrap.Environment{}, t.TempDir(), "item_keys.conf")
	if err == nil || !strings.Contains(err.Error(), "missing configuration file") {
		t.Fatalf("error = %v, want missing configuration file", err)
	}
}

func TestCollectIndexedParametersIgnoresUnrelatedVariables(t *testing.T) {
	env := bootstrap.Environment{
		"ZBX_ALLOWKEY_MY_GOOD_ENV": "value",
		"ZBX_DENYKEY_SOMETHING":    "value",
	}

	rules, err := collectIndexedParams(env, itemKeyRuleParamByPrefix)
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 0 {
		t.Fatalf("unrelated variables were parsed as item key rules: %#v", rules)
	}
}

func captureAgentStdout(t *testing.T, run func()) string {
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
