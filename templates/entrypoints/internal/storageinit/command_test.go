package storageinit_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/proxy"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/server"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/storageinit"
)

func TestStorageInitSelection(t *testing.T) {
	for _, test := range []struct {
		name, command, database        string
		historySchema, telemetrySchema string
		history                        func(bootstrap.Environment) (int, error)
	}{
		{"history", "init_db_history_providers", "history", `{"create":false}`, "invalid telemetry _SCHEMA", storageinit.PrepareHistory},
		{"telemetry", "init_db_telemetry", "telemetry", "invalid history _SCHEMA", `{"create":false}`, storageinit.PrepareHistory},
		{"telemetry without history handler", "init_db_telemetry", "telemetry", "invalid history _SCHEMA", `{"create":false}`, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			queried := false
			endpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				if !strings.Contains(string(body), "'"+test.database+"'") || !strings.HasPrefix(string(body), "SELECT") {
					t.Errorf("unexpected database access: %s", body)
				}
				queried = true
				fmt.Fprintln(w, "1")
			}))
			defer endpoint.Close()
			env := bootstrap.Environment{
				"ZBX_HISTORYPROVIDER_0":          fmt.Sprintf(`{"provider":"clickhouse","url":%q,"db":"history","types":["dbl","uint"]}`, endpoint.URL),
				"ZBX_TELEMETRYPROVIDER_0":        fmt.Sprintf(`{"provider":"clickhouse","url":%q,"db":"telemetry"}`, endpoint.URL),
				"ZBX_HISTORYPROVIDER_0_SCHEMA":   test.historySchema,
				"ZBX_TELEMETRYPROVIDER_0_SCHEMA": test.telemetrySchema,
			}
			prepareEnvironment(t, env)
			entry := bootstrap.StorageInitEntrypoint(func(bootstrap.Environment, []string) error {
				t.Fatal("database initialization invoked normal startup")
				return nil
			}, test.history, storageinit.PrepareTelemetry)
			if err := entry(env, []string{test.command}); err != nil {
				t.Fatal(err)
			}
			if !queried {
				t.Fatal("selected database was not checked")
			}
		})
	}
}

func TestSchemaTLS(t *testing.T) {
	calls := 0
	endpoint := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if len(r.TLS.PeerCertificates) != 1 {
			t.Error("client certificate was not sent")
		}
		fmt.Fprintln(w, "1")
	}))
	endpoint.TLS = &tls.Config{ClientAuth: tls.RequireAnyClientCert}
	endpoint.StartTLS()
	defer endpoint.Close()
	for _, mode := range []string{"server", "proxy", "init_db_history_providers", "init_db_telemetry"} {
		t.Run(mode, func(t *testing.T) {
			calls = 0
			env := bootstrap.Environment{}
			prepareEnvironment(t, env)
			home := env["ZABBIX_USER_HOME_DIR"]
			cert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: endpoint.Certificate().Raw})
			key, err := x509.MarshalPKCS8PrivateKey(endpoint.TLS.Certificates[0].PrivateKey)
			if err != nil {
				t.Fatal(err)
			}
			for path, data := range map[string][]byte{
				"ssl/ssl_ca_internal/ca.pem": cert,
				"ssl/certs/client.pem":       cert,
				"ssl/keys/client.key":        pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: key}),
			} {
				path = filepath.Join(home, path)
				if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, data, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			prefix := "ZBX_TELEMETRYPROVIDER_0"
			types := ""
			if mode == "init_db_history_providers" {
				prefix = "ZBX_HISTORYPROVIDER_0"
				types = `,"types":["uint"]`
			}
			env[prefix] = fmt.Sprintf(`{"provider":"clickhouse","url":%q,"db":"telemetry"%s,"ssl_cert_file":"client.pem","ssl_key_file":"client.key"}`, endpoint.URL, types)
			env[prefix+"_SCHEMA"] = `{"create":false}`
			switch mode {
			case "server":
				err = server.Prepare(env)
			case "proxy":
				err = proxy.Prepare(env, "test-proxy")
			default:
				entry := bootstrap.StorageInitEntrypoint(nil, storageinit.PrepareHistory, storageinit.PrepareTelemetry)
				err = entry(env, []string{mode})
			}
			if err != nil {
				t.Fatal(err)
			}
			if calls == 0 {
				t.Fatal("schema was not checked")
			}
			data, err := os.ReadFile(filepath.Join(env["ZBX_SSLCALOCATION"], "ca.pem"))
			if err != nil || string(data) != string(cert) {
				t.Fatalf("CA certificate was not copied: %v", err)
			}
		})
	}
}

func prepareEnvironment(t *testing.T, env bootstrap.Environment) {
	t.Helper()
	home := t.TempDir()
	env["ZABBIX_USER_HOME_DIR"] = home
	env["ZBX_SSLCALOCATION"] = filepath.Join(home, "ssl", "ssl_ca_internal")
	env["ZABBIX_CONF_DIR"] = t.TempDir()
	for _, dir := range []string{filepath.Join(home, "ssl", "ssl_ca"), env["ZBX_SSLCALOCATION"]} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"zabbix_server_modules.conf", "zabbix_proxy_modules.conf", "zabbix_server_telemetry.conf", "zabbix_proxy_telemetry.conf"} {
		if err := os.WriteFile(filepath.Join(env["ZABBIX_CONF_DIR"], name), nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestInitCancellation(t *testing.T) {
	if kind := os.Getenv("TEST_INIT_PROVIDER"); kind != "" {
		env := bootstrap.Environment{}
		prepareEnvironment(t, env)
		endpoint := os.Getenv("TEST_INIT_ENDPOINT")
		prepare := storageinit.PrepareHistory
		switch kind {
		case "elasticsearch":
			env["ZBX_HISTORYPROVIDER_0"] = fmt.Sprintf(`{"provider":"elasticsearch","url":%q,"types":["uint"]}`, endpoint)
		case "clickhouse":
			env["ZBX_HISTORYPROVIDER_0"] = fmt.Sprintf(`{"provider":"clickhouse","url":%q,"db":"history","types":["uint"]}`, endpoint)
		case "vault":
			prepare = storageinit.PrepareTelemetry
			env["ZBX_TELEMETRYPROVIDER_0"] = `{"provider":"clickhouse","url":"http://unused:8123","db":"otel","vault_path":"secret/apm"}`
			env["ZBX_VAULT"], env["ZBX_VAULTURL"], env["VAULT_TOKEN"] = "HashiCorp", endpoint, "token"
			env["ZBX_TELEMETRYPROVIDER_0_SCHEMA"] = `{"create":true}`
		}
		if kind != "vault" {
			env["ZBX_HISTORYPROVIDER_0_SCHEMA"] = `{"create":true}`
		}
		if _, err := prepare(env); !errors.Is(err, context.Canceled) {
			t.Fatalf("expected canceled initialization, got %v", err)
		}
		return
	}

	for _, kind := range []string{"clickhouse", "elasticsearch", "vault"} {
		t.Run(kind, func(t *testing.T) {
			requested := make(chan struct{}, 1)
			endpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.Copy(io.Discard, r.Body)
				requested <- struct{}{}
				<-r.Context().Done()
			}))
			defer endpoint.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestInitCancellation$")
			command.Env = append(os.Environ(), "TEST_INIT_PROVIDER="+kind, "TEST_INIT_ENDPOINT="+endpoint.URL)
			var output strings.Builder
			command.Stdout, command.Stderr = &output, &output
			if err := command.Start(); err != nil {
				t.Fatal(err)
			}
			select {
			case <-requested:
				if err := command.Process.Signal(syscall.SIGTERM); err != nil {
					t.Error(err)
				}
			case <-ctx.Done():
				t.Error("initialization did not send a request")
			}
			if err := command.Wait(); err != nil {
				t.Fatalf("initialization did not cancel cleanly: %v\n%s", err, output.String())
			}
		})
	}
}
