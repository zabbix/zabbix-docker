package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

func TestConfigureToolAccessRulesRestoresDefaults(t *testing.T) {
	templateDir := t.TempDir()
	templatePath := filepath.Join(templateDir, "tool_access.conf")
	runtimePath := filepath.Join(t.TempDir(), "tool_access.conf")
	defaults := "# Default rules\nDenyTool=item_get\nDenyTool=lld_get\n" +
		"DenyTool=interface_get\nDenyTool=macro_get\nDenyTool=history_get\n"
	if err := os.WriteFile(templatePath, []byte(defaults), 0o444); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(templateDir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(templateDir, 0o755) })

	steps := []struct {
		name string
		env  bootstrap.Environment
		want string
	}{
		{"no environment rules", bootstrap.Environment{}, defaults},
		{
			"environment overrides defaults",
			bootstrap.Environment{
				"ZBX_ALLOWTOOL_REGEXP_0": "^host_get$",
				"ZBX_DENYTOOL_1":         "*",
			},
			"# Default rules\n\nAllowToolRegexp=${ZBX_ALLOWTOOL_REGEXP_0}\nDenyTool=${ZBX_DENYTOOL_1}\n",
		},
		{"removed environment rules", bootstrap.Environment{}, defaults},
	}
	for _, step := range steps {
		t.Run(step.name, func(t *testing.T) {
			if err := ConfigureToolAccessRules(step.env, templatePath, runtimePath); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(runtimePath)
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != step.want {
				t.Fatalf("runtime rules:\n%s\nwant:\n%s", data, step.want)
			}
			data, err = os.ReadFile(templatePath)
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != defaults {
				t.Fatalf("template was changed:\n%s", data)
			}
		})
	}
}

func TestConfigureToolAccessRulesRejectsInvalidVariables(t *testing.T) {
	tests := []struct {
		name string
		env  bootstrap.Environment
	}{
		{"allow index with leading zero", bootstrap.Environment{"ZBX_ALLOWTOOL_00": "*"}},
		{"deny index with leading zero", bootstrap.Environment{"ZBX_DENYTOOL_00": "*"}},
		{"allow regexp invalid index", bootstrap.Environment{"ZBX_ALLOWTOOL_REGEXP_bad": ".*"}},
		{"deny regexp invalid index", bootstrap.Environment{"ZBX_DENYTOOL_REGEXP_bad": ".*"}},
		{"misspelled regexp prefix", bootstrap.Environment{"ZBX_DENYTOOL_REGEX_0": ".*"}},
		{"negative index", bootstrap.Environment{"ZBX_DENYTOOL_-1": "*"}},
		{"unindexed variable", bootstrap.Environment{"ZBX_DENYTOOL": "*"}},
		{"empty rule", bootstrap.Environment{"ZBX_DENYTOOL_0": ""}},
		{"missing index", bootstrap.Environment{"ZBX_DENYTOOL_1": "*"}},
		{"duplicate index", bootstrap.Environment{"ZBX_ALLOWTOOL_0": "host_get", "ZBX_DENYTOOL_0": "*"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			templatePath := filepath.Join(dir, "template.conf")
			runtimePath := filepath.Join(dir, "runtime.conf")
			original := "DenyTool=*\n"
			for _, path := range []string{templatePath, runtimePath} {
				if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			err := ConfigureToolAccessRules(test.env, templatePath, runtimePath)
			if err == nil {
				t.Fatal("invalid rule variable was accepted")
			}
			if !strings.Contains(err.Error(), "ZBX_") {
				t.Fatalf("error does not identify the invalid variables: %v", err)
			}
			for _, path := range []string{templatePath, runtimePath} {
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				if string(data) != original {
					t.Fatalf("%s was changed after invalid input: %q", path, data)
				}
			}
		})
	}
}
