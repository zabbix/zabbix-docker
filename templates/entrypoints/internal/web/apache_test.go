package web

import (
	"errors"
	"os/user"
	"testing"
)

func TestApacheRunUser(t *testing.T) {
	known := func(string) (*user.User, error) { return &user.User{Username: "zabbix"}, nil }
	unknown := func(string) (*user.User, error) { return nil, errors.New("unknown userid") }

	for _, test := range []struct {
		name   string
		uid    int
		lookup func(string) (*user.User, error)
		want   string
	}{
		{"root uses the daemon user", 0, unknown, "apache"},
		{"account with a passwd entry", 1997, known, "zabbix"},
		{"arbitrary UID without a passwd entry", 1000650000, unknown, "#1000650000"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := apacheRunUser(test.uid, "apache", test.lookup); got != test.want {
				t.Fatalf("apacheRunUser() = %q, want %q", got, test.want)
			}
		})
	}
}
