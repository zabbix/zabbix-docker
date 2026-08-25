package vault

import (
	"crypto/tls"
	"encoding/json"
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

func TestCyberArkURL(t *testing.T) {
	for _, test := range []struct {
		name   string
		prefix string
		want   string
	}{
		{
			name: "default prefix",
			want: "https://vault.example/AIMWebService/api/Accounts?AppID=zabbix",
		},
		{
			name:   "configured prefix",
			prefix: "/AIMWebService/api/Accounts?",
			want:   "https://vault.example/AIMWebService/api/Accounts?AppID=zabbix",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := cyberArkURL("https://vault.example", test.prefix, "AppID=zabbix")
			if got != test.want {
				t.Fatalf("cyberArkURL() = %q, want %q", got, test.want)
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

func TestHashiCorpAppRoleLogin(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost {
			t.Fatalf("request method = %s, want POST", req.Method)
		}
		if req.URL.String() != "https://vault.example.test/v1/auth/approle/login" {
			t.Fatalf("request URL = %s", req.URL)
		}
		if req.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("content type = %q", req.Header.Get("Content-Type"))
		}

		var payload struct {
			RoleID   string `json:"role_id"`
			SecretID string `json:"secret_id"`
		}
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if payload.RoleID != "role" || payload.SecretID != "secret" {
			t.Fatalf("unexpected AppRole payload: %#v", payload)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Body: io.NopCloser(strings.NewReader(
				`{"auth":{"client_token":"approle-token"}}`,
			)),
		}, nil
	})}

	token, err := resolveHashiCorpToken(bootstrap.Environment{
		"ZBX_VAULTAPPROLEID":   "role",
		"ZBX_VAULTAPPSECRETID": "secret",
	}, "https://vault.example.test", client)
	if err != nil {
		t.Fatal(err)
	}
	if token != "approle-token" {
		t.Fatalf("token = %q", token)
	}
}

func TestHashiCorpTokenAuthentication(t *testing.T) {
	token, err := resolveHashiCorpToken(
		bootstrap.Environment{
			"VAULT_TOKEN":          "token",
			"ZBX_VAULTAPPROLEID":   "unused-role",
			"ZBX_VAULTAPPSECRETID": "unused-secret",
		},
		"https://vault.example.test",
		&http.Client{},
	)
	if err != nil {
		t.Fatal(err)
	}
	if token != "token" {
		t.Fatalf("token = %q", token)
	}
}

func TestHashiCorpAuthenticationValidation(t *testing.T) {
	tests := []struct {
		name string
		env  bootstrap.Environment
		want string
	}{
		{
			name: "role ID without secret ID",
			env:  bootstrap.Environment{"ZBX_VAULTAPPROLEID": "role"},
			want: "ZBX_VAULTAPPSECRETID is not set",
		},
		{
			name: "secret ID without role ID",
			env:  bootstrap.Environment{"ZBX_VAULTAPPSECRETID": "secret"},
			want: "ZBX_VAULTAPPROLEID is not set",
		},
		{
			name: "no authentication",
			want: "either VAULT_TOKEN or ZBX_VAULTAPPROLEID",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := resolveHashiCorpToken(test.env, "https://vault.example.test", &http.Client{})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want containing %q", err, test.want)
			}
		})
	}
}

func TestHashiCorpAppRoleLoginRequiresToken(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Body:       io.NopCloser(strings.NewReader(`{"auth":{}}`)),
		}, nil
	})}

	_, err := loginHashiCorpAppRole(client, "https://vault.example.test", "role", "secret")
	if err == nil || !strings.Contains(err.Error(), "client token") {
		t.Fatalf("error = %v, want missing client token", err)
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

	data, err := requestWithRetry(client, http.MethodGet, "https://vault.example.test/secret", nil, func(req *http.Request) {
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
			Body:       io.NopCloser(strings.NewReader("denied: sensitive response")),
		}, nil
	})}

	_, err := requestWithRetry(client, http.MethodGet, "https://vault.example.test/secret", nil, func(*http.Request) {})
	if err != nil {
		if !strings.Contains(err.Error(), "403 Forbidden") {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Contains(err.Error(), "sensitive response") {
			t.Fatalf("Vault response body was exposed: %v", err)
		}
		return
	}
	t.Fatal("HTTP error was accepted")
}

func TestRequestWithRetryRetriesTemporaryHTTPError(t *testing.T) {
	var delays []time.Duration
	sleep = func(delay time.Duration) { delays = append(delays, delay) }
	defer func() { sleep = time.Sleep }()

	attempts := 0
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		attempts++
		if attempts == 1 {
			return &http.Response{
				StatusCode: http.StatusServiceUnavailable,
				Status:     "503 Service Unavailable",
				Body:       io.NopCloser(strings.NewReader("temporary failure")),
			}, nil
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Body:       io.NopCloser(strings.NewReader(`{"data":"ok"}`)),
		}, nil
	})}

	data, err := requestWithRetry(client, http.MethodGet, "https://vault.example.test/secret", nil, func(*http.Request) {})
	if err != nil {
		t.Fatal(err)
	}
	if attempts != 2 || len(delays) != 1 || delays[0] != retryDelay {
		t.Fatalf("attempts = %d, delays = %v", attempts, delays)
	}
	if string(data) != `{"data":"ok"}` {
		t.Fatalf("response = %s", data)
	}
}

func TestReadVaultResponseRejectsOversizedBody(t *testing.T) {
	_, err := readVaultResponse(strings.NewReader(strings.Repeat("x", maxRespSize+1)))
	if err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("oversized Vault response error = %v", err)
	}
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

	_, err := requestWithRetry(client, http.MethodGet, "https://vault.example.test/secret", nil, func(*http.Request) {})
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
