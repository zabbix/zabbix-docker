package bootstrap

import "testing"

func TestApplyDatabaseCredentials(t *testing.T) {
	tests := []struct {
		name     string
		env      Environment
		explicit bool
	}{
		{name: "environment", env: Environment{}, explicit: true},
		{
			name: "vault without database path",
			env: Environment{
				"ZBX_VAULT": "HashiCorp", "ZBX_VAULTURL": "https://vault.example.test",
			},
			explicit: true,
		},
		{
			name: "vault database credentials",
			env: Environment{
				"ZBX_VAULT": "HashiCorp", "ZBX_VAULTURL": "https://vault.example.test",
				"ZBX_VAULTDBPATH": "database/zabbix",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.env["ZBX_DB_USER"] = "old-user"
			test.env["ZBX_DB_PASSWORD"] = "old-password"

			ApplyDatabaseCredentials(test.env, "zabbix", "secret")

			if test.explicit {
				if test.env["ZBX_DB_USER"] != "zabbix" || test.env["ZBX_DB_PASSWORD"] != "secret" {
					t.Fatalf("database credentials were not applied: %#v", test.env)
				}
				return
			}

			if _, found := test.env["ZBX_DB_USER"]; found {
				t.Fatalf("database user was retained: %#v", test.env)
			}
			if _, found := test.env["ZBX_DB_PASSWORD"]; found {
				t.Fatalf("database password was retained: %#v", test.env)
			}
		})
	}
}
