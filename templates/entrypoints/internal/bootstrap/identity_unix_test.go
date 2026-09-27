//go:build !windows

package bootstrap

import (
	"errors"
	"os/user"
	"testing"
)

func TestConfigureRunUser(t *testing.T) {
	const arbitraryUID = 1000710000

	tests := []struct {
		name   string
		uid    int
		lookup func(string) (*user.User, error)
		want   Environment
	}{
		{
			name:   "root",
			uid:    0,
			lookup: func(string) (*user.User, error) { t.Fatal("root must not be looked up"); return nil, nil },
			want:   Environment{"ZBX_ALLOWROOT": "1"},
		},
		{
			name: "known user",
			uid:  1997,
			lookup: func(uid string) (*user.User, error) {
				if uid != "1997" {
					t.Fatalf("looked up %q", uid)
				}

				return &user.User{Uid: uid, Username: "zabbix"}, nil
			},
			want: Environment{"ZBX_USER": "zabbix"},
		},
		{
			// An arbitrary UID, as assigned by OpenShift, has no passwd entry.
			// Zabbix ignores User when it already runs as a non-root user, so
			// the entrypoint must keep the UID instead of failing startup.
			name:   "arbitrary UID without a passwd entry",
			uid:    arbitraryUID,
			lookup: func(string) (*user.User, error) { return nil, user.UnknownUserIdError(arbitraryUID) },
			want:   Environment{"ZBX_USER": "1000710000"},
		},
		{
			name:   "lookup failure",
			uid:    1997,
			lookup: func(string) (*user.User, error) { return nil, errors.New("passwd is unavailable") },
			want:   Environment{"ZBX_USER": "1997"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env := Environment{}
			configureRunUser(env, test.uid, test.lookup)

			if len(env) != len(test.want) {
				t.Fatalf("environment = %#v, want %#v", env, test.want)
			}
			for name, value := range test.want {
				if env[name] != value {
					t.Errorf("%s = %q, want %q", name, env[name], value)
				}
			}
		})
	}
}

func TestConfigureRunUserUsesTheCurrentProcess(t *testing.T) {
	env := Environment{}
	ConfigureRunUser(env)

	if _, root := env["ZBX_ALLOWROOT"]; root {
		if _, named := env["ZBX_USER"]; named {
			t.Fatal("both ZBX_ALLOWROOT and ZBX_USER were exported")
		}

		return
	}
	if env["ZBX_USER"] == "" {
		t.Fatalf("environment = %#v, want ZBX_USER", env)
	}
}
