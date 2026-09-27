package provider

import (
	"strings"
	"testing"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

func TestIndexedNamesIgnoresEmptyCredentialVariables(t *testing.T) {
	env := bootstrap.Environment{
		"ZBX_HISTORYPROVIDER_0":               `{"provider":"clickhouse"}`,
		"ZBX_HISTORYPROVIDER_0_USERNAME":      "",
		"ZBX_HISTORYPROVIDER_0_PASSWORD_FILE": "",
	}
	names, err := IndexedNames(env, HistoryPrefix)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 1 || names[0] != "ZBX_HISTORYPROVIDER_0" {
		t.Fatalf("unexpected names: %#v", names)
	}
}

func TestIndexedNamesRejectsNonEmptyOrphanCredential(t *testing.T) {
	_, err := IndexedNames(bootstrap.Environment{"ZBX_HISTORYPROVIDER_0_PASSWORD_FILE": "/run/secrets/password"}, HistoryPrefix)
	if err == nil || !strings.Contains(err.Error(), "matching JSON provider") {
		t.Fatalf("IndexedNames() error = %v", err)
	}
}

func TestIndexedTelemetryNamesIncludeSSLKeyPassword(t *testing.T) {
	env := bootstrap.Environment{
		"ZBX_TELEMETRYPROVIDER_0":                       `{"provider":"clickhouse"}`,
		"ZBX_TELEMETRYPROVIDER_0_SSL_KEY_PASSWORD_FILE": "/run/secrets/key-password",
	}
	names, err := IndexedNames(env, TelemetryPrefix)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 1 || names[0] != "ZBX_TELEMETRYPROVIDER_0" {
		t.Fatalf("unexpected names: %#v", names)
	}
}
