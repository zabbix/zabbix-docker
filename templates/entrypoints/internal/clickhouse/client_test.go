package clickhouse

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/testutil"
)

func TestInvalidURLDiagnostic(t *testing.T) {
	for _, address := range []string{"invalid", "ftp://clickhouse:9000", "http://clickhouse:bad"} {
		_, err := NewClient(context.Background(), Connection{URL: address})
		if err == nil || !strings.Contains(err.Error(), fmt.Sprintf("%q", address)) {
			t.Fatalf("unexpected URL error: %v", err)
		}
	}
}

func TestClientChecksHTTPAndClickHouseErrors(t *testing.T) {
	for _, status := range []int{http.StatusUnauthorized, http.StatusInternalServerError, http.StatusOK} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			calls := 0
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.URL.Query().Get("wait_end_of_query") != "1" {
					t.Error("response must be buffered")
				}
				user, password, ok := r.BasicAuth()
				if !ok || user != "user" || password != "secret-value" {
					t.Error("missing credentials")
				}
				w.WriteHeader(status)
				fmt.Fprint(w, "Code: 170. Test failure secret-value")
			})
			c.connection.Username, c.connection.Password = "user", "secret-value"
			_, err := c.query(context.Background(), "SELECT count() FROM system.databases")
			if err == nil || !strings.Contains(err.Error(), "Code: 170") || strings.Contains(err.Error(), "secret-value") {
				t.Fatalf("error = %v", err)
			}
			if calls != 1 {
				t.Fatalf("SQL/auth error retried %d times", calls)
			}
		})
	}
}

func TestClientRejectsRedirect(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "http://127.0.0.1:1/credentials")
		w.WriteHeader(http.StatusTemporaryRedirect)
	})
	_, err := c.query(context.Background(), "SELECT 1")
	if err == nil || !strings.Contains(err.Error(), "307") {
		t.Fatalf("error = %v", err)
	}
}

func TestClientQueryParameters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("max_threads") != "2" || q.Get("wait_end_of_query") != "1" || q.Get("skip_unavailable_shards") != "0" {
			t.Errorf("unexpected query parameters: %v", q)
		}
		fmt.Fprintln(w, "1")
	}))
	defer server.Close()
	client, err := NewClient(context.Background(), Connection{URL: server.URL + "?max_threads=2&wait_end_of_query=0&skip_unavailable_shards=1#section"})
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	for _, clusterDDL := range []bool{false, true} {
		if _, err := client.request(context.Background(), "SELECT 1", clusterDDL); err != nil {
			t.Fatal(err)
		}
	}
}

func TestTLSVerification(t *testing.T) {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Config.ErrorLog = log.New(io.Discard, "", 0)
	server.StartTLS()
	defer server.Close()
	dir := t.TempDir()
	cert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw})
	if err := os.WriteFile(filepath.Join(dir, "ca.pem"), cert, 0o600); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name                string
		trusted, peer, host bool
		hostname            string
		wantErr             bool
	}{
		{"both valid", true, true, true, "127.0.0.1", false},
		{"untrusted", false, true, true, "127.0.0.1", true},
		{"wrong hostname", true, true, true, "wrong.invalid", true},
		{"peer only trusted", true, true, false, "wrong.invalid", false},
		{"peer only untrusted", false, true, false, "127.0.0.1", true},
		{"host only matching", false, false, true, "127.0.0.1", false},
		{"host only mismatch", true, false, true, "wrong.invalid", true},
		{"disabled", false, false, false, "wrong.invalid", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			connection := Connection{VerifyPeer: test.peer, VerifyHost: test.host}
			if test.trusted {
				connection.CADirectory = dir
			}
			config, err := connection.tlsConfig(context.Background(), test.hostname)
			if err != nil {
				t.Fatal(err)
			}
			conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 5 * time.Second}, "tcp", server.Listener.Addr().String(), config)
			if conn != nil {
				conn.Close()
			}
			if (err != nil) != test.wantErr {
				t.Fatalf("TLS error = %v, want error = %t", err, test.wantErr)
			}
		})
	}
}

func TestCASecretMount(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "1")
	}))
	defer server.Close()
	dir := t.TempDir()
	version := filepath.Join(dir, "..2026_10_31")
	if err := os.Mkdir(version, 0o700); err != nil {
		t.Fatal(err)
	}
	cert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw})
	if err := os.WriteFile(filepath.Join(version, "ca.pem"), cert, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Base(version), filepath.Join(dir, "..data")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("..data/ca.pem", filepath.Join(dir, "ca.pem")); err != nil {
		t.Fatal(err)
	}
	invalid := filepath.Join(dir, "invalid.pem")
	if err := os.WriteFile(invalid, []byte("not a certificate"), 0o600); err != nil {
		t.Fatal(err)
	}
	var client *Client
	logs := testutil.CaptureStderr(t, func() {
		var err error
		client, err = NewClient(context.Background(), Connection{URL: server.URL, CADirectory: dir, VerifyPeer: true, VerifyHost: true})
		if err != nil {
			t.Fatal(err)
		}
	})
	defer client.Close()
	if !strings.Contains(logs, invalid) {
		t.Fatalf("missing warning for invalid CA file: %s", logs)
	}
	if _, err := client.query(context.Background(), "SELECT 1"); err != nil {
		t.Fatal(err)
	}
}

func TestReadEncryptedPrivateKey(t *testing.T) {
	openssl, err := exec.LookPath("openssl")
	if err != nil {
		t.Skip("openssl is required to test encrypted private keys")
	}
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	plainPath := filepath.Join(dir, "plain.pem")
	encryptedPath := filepath.Join(dir, "encrypted.pem")
	if err := os.WriteFile(plainPath, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}), 0o600); err != nil {
		t.Fatal(err)
	}
	const password = "test-key-password"
	command := exec.Command(openssl, "pkcs8", "-topk8", "-in", plainPath, "-out", encryptedPath, "-v2", "aes-256-cbc", "-passout", "stdin")
	command.Stdin = strings.NewReader(password + "\n")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("encrypt test key: %v: %s", err, output)
	}
	for _, test := range []struct {
		name, path, password string
		wantErr              bool
	}{
		{"correct password", encryptedPath, password, false},
		{"wrong password", encryptedPath, "wrong-password", true},
		{"missing file", filepath.Join(dir, "missing.pem"), password, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			data, err := readPrivateKey(context.Background(), test.path, test.password)
			if test.wantErr {
				var exitErr *exec.ExitError
				if !errors.As(err, &exitErr) {
					t.Fatalf("expected an OpenSSL exit error, got %v", err)
				}
				if !strings.HasPrefix(err.Error(), "openssl pkey: "+exitErr.Error()+": ") {
					t.Fatalf("OpenSSL diagnostic missing from error: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			block, _ := pem.Decode(data)
			if block == nil {
				t.Fatal("decrypted key is not PEM")
			}
			decoded, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil || !key.Equal(decoded) {
				t.Fatalf("decrypted key differs from the original: %v", err)
			}
		})
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestRequestTimeouts(t *testing.T) {
	for _, clusterDDL := range []bool{false, true} {
		t.Run(fmt.Sprint(clusterDDL), func(t *testing.T) {
			c, err := NewClient(context.Background(), Connection{URL: "http://unused:8123"})
			if err != nil {
				t.Fatal(err)
			}
			defer c.Close()
			calls := 0
			networkErr := errors.New("connection refused")
			c.http.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				deadline, ok := r.Context().Deadline()
				if !ok || time.Until(deadline) <= 0 {
					t.Error("request needs a deadline")
				}
				setting := r.URL.Query().Get("distributed_ddl_task_timeout")
				if clusterDDL {
					seconds, err := strconv.Atoi(setting)
					if err != nil || seconds <= 0 || time.Until(deadline) <= time.Duration(seconds)*time.Second {
						t.Errorf("client deadline must exceed server DDL timeout %q", setting)
					}
				} else if setting != "" {
					t.Error("ordinary request has a cluster DDL timeout")
				}
				if r.URL.Query().Get("skip_unavailable_shards") != "0" || r.URL.Query().Get("wait_end_of_query") != "1" {
					t.Error("request must set wait_end_of_query=1 and skip_unavailable_shards=0")
				}
				return nil, networkErr
			})
			_, err = c.request(context.Background(), "CREATE DATABASE `test`", clusterDDL)
			if !errors.Is(err, networkErr) || calls != 1 {
				t.Fatalf("calls=%d error=%v", calls, err)
			}
		})
	}
}
