package vault

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
	for _, test := range []struct {
		name     string
		value    string
		insecure bool
		wantErr  bool
	}{
		{name: "default verifies"},
		{name: "true", value: "true"},
		{name: "one", value: "1"},
		{name: "yes", value: "yes"},
		{name: "on", value: "on"},
		{name: "false", value: "false", insecure: true},
		{name: "zero", value: "0", insecure: true},
		{name: "no", value: "no", insecure: true},
		{name: "off", value: "off", insecure: true},
		{name: "invalid", value: "sometimes", wantErr: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			config, err := hashicorpTLSConfig(bootstrap.Environment{"ZBX_VAULT_SSL_VERIFY": test.value})
			if (err != nil) != test.wantErr {
				t.Fatalf("hashicorpTLSConfig() error = %v, wantErr = %t", err, test.wantErr)
			}
			if test.wantErr {
				return
			}
			if config.InsecureSkipVerify != test.insecure {
				t.Fatalf("InsecureSkipVerify = %t, want %t", config.InsecureSkipVerify, test.insecure)
			}
		})
	}
}

func TestHashiCorpUsesSSLCALocation(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Vault-Token") != "token" {
			t.Error("Vault token is missing")
		}
		_, _ = io.WriteString(w, `{"data":{"data":{"username":"zabbix","password":"secret"}}}`)
	}))
	defer server.Close()

	caDir := t.TempDir()
	ca := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw})
	if err := os.WriteFile(filepath.Join(caDir, "vault-ca.pem"), ca, 0o600); err != nil {
		t.Fatal(err)
	}
	credentials, err := Fetch(context.Background(), bootstrap.Environment{
		"ZBX_VAULT":            "HashiCorp",
		"ZBX_VAULTURL":         server.URL,
		"ZBX_VAULTDBPATH":      "secret/database",
		"ZBX_VAULT_SSL_VERIFY": "true",
		"ZBX_SSLCALOCATION":    caDir,
		"VAULT_TOKEN":          "token",
	})
	if err != nil {
		t.Fatal(err)
	}
	if credentials.Username != "zabbix" || credentials.Password != "secret" {
		t.Fatalf("credentials=%#v", credentials)
	}
}

func TestResolveDBCredentialsWithoutDBPath(t *testing.T) {
	creds, err := ResolveDBCredentials(bootstrap.Environment{
		"ZBX_VAULT":    "HashiCorp",
		"ZBX_VAULTURL": "https://vault.example.test",
		"VAULT_TOKEN":  "token",
	})
	if err != nil {
		t.Fatal(err)
	}
	if creds != nil {
		t.Fatalf("credentials = %#v, want nil", creds)
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

func TestDBCredentialsResolvedSecrets(t *testing.T) {
	for _, mode := range []string{"token", "approle"} {
		t.Run(mode, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/v1/auth/approle/login":
					var payload map[string]string
					if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
						t.Error(err)
					}
					if mode != "approle" || payload["role_id"] != "role" || payload["secret_id"] != "secret" {
						t.Errorf("unexpected AppRole login: %v", payload)
					}
					io.WriteString(w, `{"auth":{"client_token":"token"}}`)
				case "/v1/secret/data/database":
					if r.Header.Get("X-Vault-Token") != "token" {
						t.Error("Vault token was not supplied")
					}
					io.WriteString(w, `{"data":{"data":{"username":"zabbix","password":"password"}}}`)
				default:
					t.Errorf("unexpected request: %s", r.URL.Path)
					http.NotFound(w, r)
				}
			}))
			defer server.Close()
			env := bootstrap.Environment{
				"ZBX_VAULT": "HashiCorp", "ZBX_VAULTURL": server.URL, "ZBX_VAULTDBPATH": "secret/database",
			}
			secrets := map[string]string{"VAULT_TOKEN": "token"}
			if mode == "approle" {
				secrets = map[string]string{"ZBX_VAULTAPPROLEID": "role", "ZBX_VAULTAPPSECRETID": "secret"}
			}
			for name, value := range secrets {
				env[name] = value
			}
			creds, err := ResolveDBCredentials(env)
			if err != nil {
				t.Fatal(err)
			}
			if creds == nil || *creds != (Credentials{Username: "zabbix", Password: "password"}) {
				t.Fatalf("unexpected credentials: %v", creds)
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

	token, err := resolveHashiCorpToken(context.Background(), bootstrap.Environment{
		"ZBX_VAULTAPPROLEID":   "role",
		"ZBX_VAULTAPPSECRETID": "secret",
	}, "https://vault.example.test", testVaultClient(client, nil))
	if err != nil {
		t.Fatal(err)
	}
	if token != "approle-token" {
		t.Fatalf("token = %q", token)
	}
}

func TestHashiCorpTokenAuthentication(t *testing.T) {
	token, err := resolveHashiCorpToken(context.Background(),
		bootstrap.Environment{
			"VAULT_TOKEN":          "token",
			"ZBX_VAULTAPPROLEID":   "unused-role",
			"ZBX_VAULTAPPSECRETID": "unused-secret",
		},
		"https://vault.example.test",
		testVaultClient(&http.Client{}, nil),
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
			_, err := resolveHashiCorpToken(context.Background(), test.env, "https://vault.example.test", testVaultClient(&http.Client{}, nil))
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

	_, err := loginHashiCorpAppRole(context.Background(), testVaultClient(client, nil), "https://vault.example.test", "role", "secret")
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

	data, err := testVaultClient(client, nil).request(context.Background(), http.MethodGet, "https://vault.example.test/secret", nil, func(req *http.Request) {
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

	_, err := testVaultClient(client, nil).request(context.Background(), http.MethodGet, "https://vault.example.test/secret", nil, func(*http.Request) {})
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
	recordDelay := func(_ context.Context, delay time.Duration) error {
		delays = append(delays, delay)

		return nil
	}

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

	data, err := testVaultClient(client, recordDelay).request(context.Background(), http.MethodGet, "https://vault.example.test/secret", nil, func(*http.Request) {})
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
	countSleep := func(context.Context, time.Duration) error {
		sleeps++

		return nil
	}

	attempts := 0
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		attempts++
		return nil, errors.New("connection refused")
	})}

	_, err := testVaultClient(client, countSleep).request(context.Background(), http.MethodGet, "https://vault.example.test/secret", nil, func(*http.Request) {})
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

// testVaultClient uses the production retry policy with the given wait
// function, so that tests observe the pauses instead of sleeping.
func testVaultClient(httpClient *http.Client, wait func(context.Context, time.Duration) error) vaultClient {
	if wait == nil {
		wait = func(context.Context, time.Duration) error { return nil }
	}

	return vaultClient{http: httpClient, wait: wait}
}
