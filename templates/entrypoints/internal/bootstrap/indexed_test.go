package bootstrap

import (
	"strings"
	"testing"
)

func TestCollectIndexedUsesTheLongestPrefix(t *testing.T) {
	env := Environment{
		"ZBX_DENYKEY_0":        "system.run[*]",
		"ZBX_DENYKEY_REGEXP_1": "^system\\.run",
	}
	variables, err := CollectIndexed(env, []string{"ZBX_DENYKEY", "ZBX_DENYKEY_REGEXP"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(variables) != 2 || variables[1].Prefix != "ZBX_DENYKEY_REGEXP" {
		t.Fatalf("indexed variables = %#v", variables)
	}
}

func TestCollectIndexedValidation(t *testing.T) {
	tests := []struct {
		name   string
		env    Environment
		strict bool
		want   string
	}{
		{name: "unindexed", env: Environment{"ZBX_ALIAS": "value"}, want: "ZBX_ALIAS is not supported"},
		{name: "empty", env: Environment{"ZBX_ALIAS_0": ""}, want: "ZBX_ALIAS_0 must not be empty"},
		{name: "gap", env: Environment{"ZBX_ALIAS_1": "value"}, want: "index 0 is missing"},
		{name: "strict suffix", env: Environment{"ZBX_ALIAS_BAD": "value"}, strict: true,
			want: "invalid indexed variable ZBX_ALIAS_BAD"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := CollectIndexed(test.env, []string{"ZBX_ALIAS"}, test.strict)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want containing %q", err, test.want)
			}
		})
	}
}

func TestCollectIndexedIgnoresUnknownSuffix(t *testing.T) {
	variables, err := CollectIndexed(Environment{"ZBX_ALIAS_DESCRIPTION": "value"}, []string{"ZBX_ALIAS"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(variables) != 0 {
		t.Fatalf("indexed variables = %#v", variables)
	}
}
