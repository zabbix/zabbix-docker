package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

func TestMergeParameterValues(t *testing.T) {
	for _, test := range []struct{ name, input, value, want string }{
		{
			name:  "quoted values",
			input: "# DenyKey=system.run[*]\nOther=value\n",
			value: `"one,two"`,
			want:  "# DenyKey=system.run[*]\nDenyKey=one\nDenyKey=two\nOther=value\n",
		},
		{
			name:  "CRLF",
			input: "# DenyKey=system.run[*]\r\nOther=value\r\n",
			value: "one,two",
			want:  "# DenyKey=system.run[*]\nDenyKey=one\nDenyKey=two\nOther=value\n",
		},
		{
			name:  "existing values",
			input: "DenyKey=existing\n# DenyKey=system.run[*]\n",
			value: "existing,new",
			want:  "DenyKey=existing\nDenyKey=new\n# DenyKey=system.run[*]\n",
		},
		{
			name:  "empty value",
			input: "DenyKey=system.run[*]\nOther=value\n",
			want:  "Other=value\n",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "agent.conf")
			if err := os.WriteFile(path, []byte(test.input), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := MergeParameterValues(path, "DenyKey", test.value); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != test.want {
				t.Fatalf("config:\n%q\nwant:\n%q", data, test.want)
			}
			info, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			// Windows does not expose POSIX owner/group/other permission bits.
			if runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
				t.Fatalf("config mode = %o, want 600", info.Mode().Perm())
			}
		})
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

func TestUpdateIndexedParameter(t *testing.T) {
	path := filepath.Join(t.TempDir(), "server.conf")
	if err := os.WriteFile(path, []byte("# HistoryProvider=\nOther=value\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	env := bootstrap.Environment{
		"ZBX_HISTORYPROVIDER_0": "provider-one",
		"ZBX_HISTORYPROVIDER_1": "provider-two",
	}
	if err := UpdateIndexedParameter(env, path, "HistoryProvider", "ZBX_HISTORYPROVIDER"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := "# HistoryProvider=\n" +
		"HistoryProvider=${ZBX_HISTORYPROVIDER_0}\n" +
		"HistoryProvider=${ZBX_HISTORYPROVIDER_1}\n" +
		"Other=value\n"
	if string(data) != want {
		t.Fatalf("config:\n%s\nwant:\n%s", data, want)
	}
	if env["ZBX_HISTORYPROVIDER_0"] == "" || env["ZBX_HISTORYPROVIDER_1"] == "" {
		t.Fatal("referenced indexed variables were removed")
	}
}

func TestUpdateIndexedParameterValidation(t *testing.T) {
	tests := []struct {
		name string
		env  bootstrap.Environment
		want string
	}{
		{
			name: "legacy variable",
			env:  bootstrap.Environment{"ZBX_ALIAS": "custom:system.hostname"},
			want: "ZBX_ALIAS is not supported",
		},
		{
			name: "empty legacy variable",
			env:  bootstrap.Environment{"ZBX_ALIAS": ""},
			want: "ZBX_ALIAS is not supported",
		},
		{
			name: "empty indexed variable",
			env:  bootstrap.Environment{"ZBX_ALIAS_0": ""},
			want: "ZBX_ALIAS_0 must not be empty",
		},
		{
			name: "missing index",
			env: bootstrap.Environment{
				"ZBX_ALIAS_0": "custom.one:system.hostname",
				"ZBX_ALIAS_2": "custom.two:system.uname",
			},
			want: "ZBX_ALIAS_2 uses index 2, but index 1 is missing",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := UpdateIndexedParameter(test.env, "unused.conf", "Alias", "ZBX_ALIAS")
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want containing %q", err, test.want)
			}
		})
	}
}
