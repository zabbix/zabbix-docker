package postgresql

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

func TestParseEndpoints(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		port    string
		want    []endpoint
		wantErr string
	}{
		{
			name: "default Unix socket",
			host: "",
			port: "5432",
			want: []endpoint{{host: socketDir, port: "5432"}},
		},
		{
			name: "single hostname",
			host: "postgres-server",
			port: "5432",
			want: []endpoint{{host: "postgres-server", port: "5432"}},
		},
		{
			name: "mixed endpoints",
			host: "postgres-1:5433, postgres-2, [2001:db8::1]:5434,2001:db8::2",
			port: "5432",
			want: []endpoint{
				{host: "postgres-1", port: "5433"},
				{host: "postgres-2", port: "5432"},
				{host: "2001:db8::1", port: "5434"},
				{host: "2001:db8::2", port: "5432"},
			},
		},
		{
			name: "Zabbix documentation example",
			host: "localhost:5431,127.0.0.1:20051,zabbix.domain,[::1]:30051,[12fc::1]",
			port: "5432",
			want: []endpoint{
				{host: "localhost", port: "5431"},
				{host: "127.0.0.1", port: "20051"},
				{host: "zabbix.domain", port: "5432"},
				{host: "::1", port: "30051"},
				{host: "12fc::1", port: "5432"},
			},
		},
		{
			name: "Unix socket directory",
			host: "/var/run/pgbouncer",
			port: "6432",
			want: []endpoint{{host: "/var/run/pgbouncer", port: "6432"}},
		},
		{
			name: "whitespace around ports",
			host: "postgres-1 : 5432, [ 2001:db8::1 ] : 5433",
			port: "5432",
			want: []endpoint{
				{host: "postgres-1", port: "5432"},
				{host: "2001:db8::1", port: "5433"},
			},
		},
		{name: "empty endpoint", host: "postgres-1,,postgres-2", port: "5432", wantErr: "endpoint 2 is empty"},
		{name: "missing IPv6 bracket", host: "[2001:db8::1", port: "5432", wantErr: "missing closing bracket"},
		{name: "IPv6 suffix", host: "[2001:db8::1]extra", port: "5432", wantErr: "unexpected data"},
		{name: "empty host", host: ":5432", port: "5432", wantErr: "host is empty"},
		{name: "empty port", host: "postgres-1:", port: "5432", wantErr: "port is empty"},
		{name: "invalid port", host: "postgres-1:postgresql", port: "5432", wantErr: "invalid port"},
		{name: "port overflow", host: "postgres-1:65536", port: "5432", wantErr: "invalid port"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseEndpoints(test.host, test.port)
			if test.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantErr) {
					t.Fatalf("parseEndpoints() error = %v, want error containing %q", err, test.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("parseEndpoints() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestMultipleEndpointsConnectionConfig(t *testing.T) {
	hosts := "postgres-1:5433,postgres-2,[2001:db8::1]:5434,2001:db8::2"
	env := bootstrap.Environment{
		"DB_SERVER_HOST":   hosts,
		"DB_SERVER_PORT":   "5432",
		"ZBX_DBTLSCONNECT": "verify_full",
	}
	db := NewForBackend(env)
	if err := db.Configure("zabbix"); err != nil {
		t.Fatal(err)
	}

	config, err := db.connConfig("zabbix", "zabbix", "secret")
	if err != nil {
		t.Fatal(err)
	}

	type target struct {
		host string
		port uint16
	}
	got := []target{
		{host: config.Host, port: config.Port},
	}
	for _, fallback := range config.Fallbacks {
		got = append(got, target{host: fallback.Host, port: fallback.Port})
	}
	want := []target{
		{host: "postgres-1", port: 5433},
		{host: "postgres-2", port: 5432},
		{host: "2001:db8::1", port: 5434},
		{host: "2001:db8::2", port: 5432},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("PostgreSQL connection targets = %#v, want %#v", got, want)
	}
	wantTimeout := time.Duration(connectTimeout) * time.Second
	if config.ConnectTimeout != wantTimeout {
		t.Fatalf("ConnectTimeout = %s, want %s", config.ConnectTimeout, wantTimeout)
	}
	if config.ValidateConnect == nil {
		t.Fatal("multiple endpoints do not require a read-write connection")
	}
	if config.TLSConfig == nil || config.TLSConfig.ServerName != "postgres-1" {
		t.Fatalf("primary TLS server name = %q, want postgres-1", config.TLSConfig.ServerName)
	}
	for i, fallback := range config.Fallbacks {
		if fallback.TLSConfig == nil || fallback.TLSConfig.ServerName != want[i+1].host {
			t.Fatalf("fallback %d TLS server name is not configured for %q", i, want[i+1].host)
		}
	}

	db.ExportEnv()
	if env["ZBX_DB_HOST"] != hosts || env["ZBX_DB_PORT"] != "5432" {
		t.Fatalf("exported target was normalized unexpectedly: host=%q port=%q", env["ZBX_DB_HOST"], env["ZBX_DB_PORT"])
	}
}

func TestSingleEndpointConnectionConfig(t *testing.T) {
	db := NewForBackend(bootstrap.Environment{"DB_SERVER_HOST": "postgres-server"})
	if err := db.Configure("zabbix"); err != nil {
		t.Fatal(err)
	}

	config, err := db.connConfig("zabbix", "zabbix", "secret")
	if err != nil {
		t.Fatal(err)
	}
	if config.Host != "postgres-server" || config.Port != 5432 || len(config.Fallbacks) != 0 {
		t.Fatalf("unexpected single-endpoint config: host=%q port=%d fallbacks=%d", config.Host, config.Port, len(config.Fallbacks))
	}
	wantTimeout := time.Duration(connectTimeout) * time.Second
	if config.ConnectTimeout != wantTimeout {
		t.Fatalf("ConnectTimeout = %s, want %s", config.ConnectTimeout, wantTimeout)
	}
	if config.ValidateConnect != nil {
		t.Fatal("single endpoint unexpectedly requires a read-write connection")
	}
}

func TestUnixSocketConnectionConfig(t *testing.T) {
	db := NewForBackend(bootstrap.Environment{"DB_SERVER_HOST": ""})
	if err := db.Configure("zabbix"); err != nil {
		t.Fatal(err)
	}

	config, err := db.connConfig("zabbix", "zabbix", "secret")
	if err != nil {
		t.Fatal(err)
	}
	if config.Host != socketDir || config.Port != 5432 || len(config.Fallbacks) != 0 {
		t.Fatalf("unexpected Unix socket config: host=%q port=%d fallbacks=%d", config.Host, config.Port, len(config.Fallbacks))
	}
}
