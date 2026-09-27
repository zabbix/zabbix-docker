package provider

import (
	"strings"
	"testing"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

// The image only checks that the value types can be rendered; the service
// validates the names.
func TestValueTypesAreNotRestricted(t *testing.T) {
	for _, value := range []string{
		`{"provider":"elasticsearch","types":["uint",""],"url":"http://elasticsearch:9200"}`,
		"{\"provider\":\"elasticsearch\",\"types\":[\"uint\",\"future\\nLogFile=/tmp/zabbix.log\"],\"url\":\"http://elasticsearch:9200\"}",
	} {
		env := bootstrap.Environment{"ZBX_HISTORYPROVIDER_0": value}
		err := ConfigureNative(env, History)
		if err == nil || !strings.Contains(err.Error(), "must contain value type names") {
			t.Fatalf("error = %v", err)
		}
	}

	env := bootstrap.Environment{
		"ZBX_HISTORYPROVIDER_0": `{"provider":"elasticsearch","types":["uint","future"],"url":"http://elasticsearch:9200"}`,
	}
	if err := ConfigureNative(env, History); err != nil {
		t.Fatalf("unknown value type was rejected: %v", err)
	}
}

func TestUnsupportedProvidersAreRejected(t *testing.T) {
	for _, test := range []struct {
		param *parameter
		value string
	}{
		{History, `{"provider":"unknown","types":["uint"],"url":"http://example:9000"}`},
		{Telemetry, `{"provider":"elasticsearch","url":"http://example:9200","db":"zabbix"}`},
	} {
		env := bootstrap.Environment{bootstrap.IndexedName(test.param.Prefix, 0): test.value}
		if err := ConfigureNative(env, test.param); err == nil {
			t.Errorf("%s accepted an unsupported provider", test.param.Prefix)
		}
	}
}

// source_ip and log_slow_queries are documented per-provider options, so the
// image renders them instead of rejecting them; only the frontend, which does
// not use them, leaves them out.
func TestPerProviderConnectionOptionsReachTheServer(t *testing.T) {
	value := `{"provider":"clickhouse","types":["uint"],"url":"http://clickhouse:8123","db":"zabbix","source_ip":"192.0.2.1","log_slow_queries":3000}`
	env := bootstrap.Environment{"ZBX_HISTORYPROVIDER_0": value}
	if err := ConfigureNative(env, History); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{`source_ip="192.0.2.1"`, "log_slow_queries=3000"} {
		if !strings.Contains(env["ZBX_HISTORYPROVIDER_0"], expected) {
			t.Errorf("provider %q does not contain %q", env["ZBX_HISTORYPROVIDER_0"], expected)
		}
	}

	web := bootstrap.Environment{"ZBX_HISTORYPROVIDER_0": value}
	providers, err := WebProviders(web, History)
	if err != nil {
		t.Fatal(err)
	}
	for _, serverOnly := range []string{"source_ip", "log_slow_queries"} {
		if _, exists := providers[0][serverOnly]; exists {
			t.Errorf("%s reached the frontend", serverOnly)
		}
	}
}

// The directories of the certificate store are prepared by the image, so a
// provider cannot redirect them.
func TestManagedLocationsAreRejectedForEveryParameter(t *testing.T) {
	for _, test := range []struct {
		param *parameter
		value string
	}{
		{param: History, value: `{"provider":"clickhouse","types":["uint"],"url":"http://clickhouse:8123","ssl_ca_location":"/ca"}`},
		{param: Telemetry, value: `{"provider":"clickhouse","url":"https://clickhouse:8443","db":"zabbix","ssl_ca_location":"/ca"}`},
	} {
		env := bootstrap.Environment{bootstrap.IndexedName(test.param.Prefix, 0): test.value}
		err := ConfigureNative(env, test.param)
		if err == nil || !strings.Contains(err.Error(), "managed by the image") {
			t.Errorf("%s error = %v", test.param.Prefix, err)
		}
	}
}
