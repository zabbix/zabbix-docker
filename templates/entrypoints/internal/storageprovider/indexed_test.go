package storageprovider

import (
	"strings"
	"testing"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

func TestIndexedNamesIgnoresEmptyCredentialVariables(t *testing.T) {
	env := bootstrap.Environment{
		"ZBX_HISTORYPROVIDER_0":          `{"provider":"clickhouse"}`,
		"ZBX_HISTORYPROVIDER_0_USERNAME": "",
		"ZBX_HISTORYPROVIDER_0_PASSWORD": "",
	}
	names, err := indexedNames(env, historyPrefix)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 1 || names[0] != "ZBX_HISTORYPROVIDER_0" {
		t.Fatalf("unexpected names: %#v", names)
	}
}

func TestIndexedNamesRejectsNonEmptyOrphanCredential(t *testing.T) {
	for _, variable := range []string{
		"ZBX_HISTORYPROVIDER_0_PASSWORD",
		"ZBX_HISTORYPROVIDER_0_USERNAME_FILE",
		"ZBX_HISTORYPROVIDER_0_SSL_CALOCATION",
	} {
		t.Run(variable, func(t *testing.T) {
			_, err := indexedNames(bootstrap.Environment{variable: "secret"}, historyPrefix)
			if err == nil || !strings.Contains(err.Error(), "matching JSON provider") {
				t.Fatalf("indexedNames() error = %v", err)
			}
		})
	}
}

func TestIndexedTelemetryNamesIncludeSSLKeyPassword(t *testing.T) {
	env := bootstrap.Environment{
		"ZBX_TELEMETRYPROVIDER_0":                  `{"provider":"clickhouse"}`,
		"ZBX_TELEMETRYPROVIDER_0_SSL_KEY_PASSWORD": "secret",
	}
	names, err := indexedNames(env, telemetryPrefix)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 1 || names[0] != "ZBX_TELEMETRYPROVIDER_0" {
		t.Fatalf("unexpected names: %#v", names)
	}
}

func TestCredentialBaseRecognizesOnlyCALocation(t *testing.T) {
	for _, suffix := range []string{"_SSL_CERTLOCATION", "_SSL_KEYLOCATION"} {
		if _, found := credentialBase("ZBX_HISTORYPROVIDER_0" + suffix); found {
			t.Errorf("credentialBase() accepted unsupported suffix %s", suffix)
		}
	}
	if base, found := credentialBase("ZBX_HISTORYPROVIDER_0_SSL_CALOCATION"); !found || base != "ZBX_HISTORYPROVIDER_0" {
		t.Fatalf("credentialBase() = %q, %t", base, found)
	}
}
