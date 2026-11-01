package storageprovider

import (
	"fmt"
	"slices"
	"testing"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

func TestConfigHistoryValueTypes(t *testing.T) {
	for _, provider := range []string{"clickhouse", "elasticsearch"} {
		for _, test := range []struct {
			types string
			valid bool
		}{
			{`["uint","future","uint"]`, true},
			{`null`, false},
			{`[]`, false},
			{`"uint"`, false},
			{`[null]`, false},
			{`[42]`, false},
			{`[""]`, false},
			{`["uint\n"]`, false},
		} {
			t.Run(provider+"/"+test.types, func(t *testing.T) {
				raw := fmt.Sprintf(`{"provider":%q,"types":%s}`, provider, test.types)
				env := bootstrap.Environment{
					"ZBX_HISTORYPROVIDER_0":        raw,
					"ZBX_HISTORYPROVIDER_0_SCHEMA": `{"create":false}`,
				}
				configuration, err := NewConfiguration(env, History)
				if err != nil {
					t.Fatal(err)
				}
				configs, err := configuration.ConfigsToPrepare()
				if !test.valid {
					if err == nil {
						t.Fatal("invalid history types accepted")
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				if len(configs) != 1 || !slices.Equal(configs[0].ValueTypes, []string{"uint", "future", "uint"}) {
					t.Fatalf("types or their order changed: %+v", configs)
				}
				if env["ZBX_HISTORYPROVIDER_0"] != raw {
					t.Fatal("schema parsing changed the service configuration")
				}
			})
		}
	}
}

func TestFrontendIgnoresSchemaOptions(t *testing.T) {
	env := bootstrap.Environment{
		"ZBX_TELEMETRYPROVIDER_0":        `{"provider":"clickhouse","url":"http://unused:8123","db":"data"}`,
		"ZBX_TELEMETRYPROVIDER_0_SCHEMA": `{"create":true,"cluster":"test"}`,
	}
	providers, err := WebProviders(env, Telemetry)
	if err != nil {
		t.Fatal(err)
	}
	if len(providers) != 1 {
		t.Fatal(providers)
	}
	if _, exists := env["ZBX_TELEMETRYPROVIDER_0_SCHEMA"]; exists {
		t.Fatal("_SCHEMA was not removed from the frontend environment")
	}
	for _, key := range []string{"schema", "cluster", "create"} {
		if _, exists := providers[0][key]; exists {
			t.Fatalf("frontend received %s", key)
		}
	}
}

func TestSchemaOptionsRequireProvider(t *testing.T) {
	_, err := indexedNames(bootstrap.Environment{"ZBX_HISTORYPROVIDER_0_SCHEMA": `{"create":false}`}, historyPrefix)
	if err == nil {
		t.Fatal("_SCHEMA accepted without a provider")
	}
}
