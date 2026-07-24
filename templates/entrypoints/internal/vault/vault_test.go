package vault

import (
	"crypto/tls"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestHashiCorpURL(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		path   string
		want   string
	}{
		{"mount and secret", "", "team/zabbix", "https://vault.example/v1/team/data/zabbix"},
		{"nested secret path", "", "secret/team/zabbix", "https://vault.example/v1/secret/data/team/zabbix"},
		{"no mount point", "", "zabbix", "https://vault.example/v1/zabbix/data/zabbix"},
		{"explicit prefix", "/custom/data/", "team/zabbix", "https://vault.example/custom/data/team/zabbix"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := hashicorpURL("https://vault.example", tc.prefix, tc.path); got != tc.want {
				t.Fatalf("hashicorpURL(%q, %q) = %q, want %q", tc.prefix, tc.path, got, tc.want)
			}
		})
	}
}

func TestHashiCorpTLSConfig(t *testing.T) {
	if !hashicorpTLSConfig(bootstrap.Environment{}).InsecureSkipVerify {
		t.Fatal("certificate verification must stay disabled by default")
	}
	if hashicorpTLSConfig(bootstrap.Environment{"ZBX_VAULT_SSL_VERIFY": "true"}).InsecureSkipVerify {
		t.Fatal("ZBX_VAULT_SSL_VERIFY=true did not enable certificate verification")
	}
	if !hashicorpTLSConfig(bootstrap.Environment{"ZBX_VAULT_SSL_VERIFY": "false"}).InsecureSkipVerify {
		t.Fatal("ZBX_VAULT_SSL_VERIFY=false must keep verification disabled")
	}
}

func TestDecodeHashiCorp(t *testing.T) {
	for _, test := range []struct {
		name      string
		response  string
		wantPass  string
		wantError string
	}{
		{
			name:     "credentials",
			response: `{"data":{"data":{"username":"zabbix","password":"secret"}}}`,
			wantPass: "secret",
		},
		{
			name:     "empty password",
			response: `{"data":{"data":{"username":"zabbix","password":""}}}`,
		},
		{
			name:      "missing username",
			response:  `{"data":{"data":{"password":"secret"}}}`,
			wantError: "database username",
		},
		{
			name:      "empty username",
			response:  `{"data":{"data":{"username":"","password":"secret"}}}`,
			wantError: "database username",
		},
		{
			name:      "missing password",
			response:  `{"data":{"data":{"username":"zabbix"}}}`,
			wantError: "database password",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			creds, err := decodeHashiCorp([]byte(test.response))
			if test.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantError) {
					t.Fatalf("decodeHashiCorp() error = %v, want error containing %q", err, test.wantError)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if creds.Username != "zabbix" || creds.Password != test.wantPass {
				t.Fatalf("unexpected credentials: %#v", creds)
			}
		})
	}
}

func TestDecodeCyberArk(t *testing.T) {
	for _, test := range []struct {
		name      string
		response  string
		wantPass  string
		wantError string
	}{
		{
			name:     "credentials",
			response: `{"UserName":"zabbix","Content":"secret"}`,
			wantPass: "secret",
		},
		{
			name:     "empty password",
			response: `{"UserName":"zabbix","Content":""}`,
		},
		{
			name:      "missing username",
			response:  `{"Content":"secret"}`,
			wantError: "database username",
		},
		{
			name:      "empty username",
			response:  `{"UserName":"","Content":"secret"}`,
			wantError: "database username",
		},
		{
			name:      "missing password",
			response:  `{"UserName":"zabbix"}`,
			wantError: "database password",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			creds, err := decodeCyberArk([]byte(test.response))
			if test.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantError) {
					t.Fatalf("decodeCyberArk() error = %v, want error containing %q", err, test.wantError)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if creds.Username != "zabbix" || creds.Password != test.wantPass {
				t.Fatalf("unexpected credentials: %#v", creds)
			}
		})
	}
}

func TestRequestWithRetry(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Header.Get("X-Vault-Token") != "token" {
			t.Fatalf("unexpected Vault token: %q", req.Header.Get("X-Vault-Token"))
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Body:       io.NopCloser(strings.NewReader(`{"data":"ok"}`)),
		}, nil
	})}

	data, err := requestWithRetry(client, "https://vault.example.test/secret", func(req *http.Request) {
		req.Header.Set("X-Vault-Token", "token")
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"data":"ok"}` {
		t.Fatalf("response = %s", data)
	}
}

func TestRequestWithRetryRejectsHTTPError(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusForbidden,
			Status:     "403 Forbidden",
			Body:       io.NopCloser(strings.NewReader("denied")),
		}, nil
	})}

	_, err := requestWithRetry(client, "https://vault.example.test/secret", func(*http.Request) {})
	if err != nil {
		if !strings.Contains(err.Error(), "403 Forbidden") {
			t.Fatalf("unexpected error: %v", err)
		}
		return
	}
	t.Fatal("HTTP error was accepted")
}

func TestRequestWithRetryGivesUp(t *testing.T) {
	sleeps := 0
	sleep = func(time.Duration) { sleeps++ }
	defer func() { sleep = time.Sleep }()

	attempts := 0
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		attempts++
		return nil, errors.New("connection refused")
	})}

	_, err := requestWithRetry(client, "https://vault.example.test/secret", func(*http.Request) {})
	if err == nil {
		t.Fatal("unavailable vault did not result in an error")
	}
	if !strings.Contains(err.Error(), "connection refused") {
		t.Fatalf("error does not mention the cause: %v", err)
	}
	if attempts != 3 {
		t.Fatalf("attempts = %d, want %d", attempts, 3)
	}
	if sleeps != 2 {
		t.Fatalf("sleeps = %d, want %d", sleeps, 2)
	}
}

func TestVaultTransportPreservesProxySupport(t *testing.T) {
	tlsConfig := &tls.Config{ServerName: "vault.example.test"}
	transport := vaultTransport(tlsConfig)
	if transport.Proxy == nil {
		t.Fatal("Vault transport does not use proxy environment variables")
	}
	if transport.TLSClientConfig != tlsConfig {
		t.Fatal("Vault TLS configuration was not applied")
	}
}
