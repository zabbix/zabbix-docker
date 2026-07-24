// Package vault fetches database credentials from HashiCorp Vault or
// CyberArk, mirroring the Vault support of Zabbix server itself.
package vault

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
)

// Credentials contains a database username and password fetched from Vault.
type Credentials struct {
	Username string
	Password string
}

// ResolveDBCredentials fetches the database credentials when
// ZBX_VAULT is configured. It returns nil without an error when no Vault is
// in use.
func ResolveDBCredentials(env bootstrap.Environment) (*Credentials, error) {
	if env["ZBX_VAULT"] == "" {
		return nil, nil
	}

	bootstrap.LogInfo("***** Connecting to vault... ******")

	creds, err := Fetch(env)
	if err != nil {
		return nil, err
	}

	return &creds, nil
}

// Fetch retrieves the credentials from the Vault selected by ZBX_VAULT.
func Fetch(env bootstrap.Environment) (Credentials, error) {
	vaultURL := strings.TrimSuffix(env["ZBX_VAULTURL"], "/")
	dbPath := strings.Trim(env["ZBX_VAULTDBPATH"], "/")

	if vaultURL == "" || dbPath == "" {
		return Credentials{}, fmt.Errorf("if ZBX_VAULT is used, ZBX_VAULTURL and ZBX_VAULTDBPATH must be set")
	}

	switch env["ZBX_VAULT"] {
	case "HashiCorp":
		return fetchHashiCorp(env, vaultURL, dbPath)
	case "CyberArk":
		return fetchCyberArk(env, vaultURL, dbPath)
	default:
		return Credentials{}, fmt.Errorf("ZBX_VAULT has wrong value: HashiCorp or CyberArk are supported")
	}
}

func fetchHashiCorp(env bootstrap.Environment, baseURL, dbPath string) (Credentials, error) {
	reqURL := hashicorpURL(baseURL, env["ZBX_VAULTPREFIX"], dbPath)

	bootstrap.LogInfo("***** VAULT URL: %s", reqURL)

	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: vaultTransport(hashicorpTLSConfig(env)),
	}

	data, err := requestWithRetry(client, reqURL, func(req *http.Request) {
		req.Header.Set("X-Vault-Token", env["VAULT_TOKEN"])
	})
	if err != nil {
		return Credentials{}, err
	}

	return decodeHashiCorp(data)
}

func hashicorpTLSConfig(env bootstrap.Environment) *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: env["ZBX_VAULT_SSL_VERIFY"] != "true",
	}
}

func hashicorpURL(baseURL, configuredPrefix, dbPath string) string {
	prefix := strings.Trim(configuredPrefix, "/")
	if prefix == "" {
		// KV v2 API expects "data" right after the mount point (the first
		// path segment), the same way Zabbix server resolves its Vault paths.
		mount, secretPath, found := strings.Cut(dbPath, "/")
		if !found {
			// No mount point in the path; keep the legacy shell entrypoint
			// behaviour of repeating the single segment.
			secretPath = dbPath
		}
		prefix = "v1/" + mount + "/data"
		dbPath = secretPath
	}

	return baseURL + "/" + prefix + "/" + dbPath
}

func decodeHashiCorp(data []byte) (Credentials, error) {
	var resp struct {
		Errors []string `json:"errors"`
		Data   struct {
			Data struct {
				Username *string `json:"username"`
				Password *string `json:"password"`
			} `json:"data"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return Credentials{}, fmt.Errorf("decode HashiCorp Vault response: %w", err)
	}

	if len(resp.Errors) != 0 {
		return Credentials{}, fmt.Errorf("error getting secrets from vault: %s", strings.Join(resp.Errors, ", "))
	}
	if resp.Data.Data.Username == nil || *resp.Data.Data.Username == "" {
		return Credentials{}, fmt.Errorf("vault response from HashiCorp does not contain a database username")
	}
	if resp.Data.Data.Password == nil {
		return Credentials{}, fmt.Errorf("vault response from HashiCorp does not contain a database password")
	}

	return Credentials{Username: *resp.Data.Data.Username, Password: *resp.Data.Data.Password}, nil
}

func fetchCyberArk(env bootstrap.Environment, baseURL, dbPath string) (Credentials, error) {
	certFile := env["ZBX_VAULTCERTFILE"]
	if certFile == "" {
		certFile = env["ZBX_VAULTTLSCERTFILE"]
	}
	if certFile == "" {
		return Credentials{}, fmt.Errorf("if CyberArk is used, a Vault certificate file must be set")
	}

	certPEM, err := os.ReadFile(certFile)
	if err != nil {
		return Credentials{}, fmt.Errorf("read CyberArk certificate: %w", err)
	}

	keyPEM := certPEM
	keyFile := env["ZBX_VAULTKEYFILE"]
	if keyFile == "" {
		keyFile = env["ZBX_VAULTTLSKEYFILE"]
	}
	if keyFile != "" {
		keyPEM, err = os.ReadFile(keyFile)
		if err != nil {
			return Credentials{}, fmt.Errorf("read CyberArk private key: %w", err)
		}
	}

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return Credentials{}, fmt.Errorf("load CyberArk client certificate: %w", err)
	}

	prefix := env["ZBX_VAULTPREFIX"]
	if prefix == "" {
		prefix = "AIMWebService/api/Accounts?"
	} else {
		prefix = strings.Trim(prefix, "/") + "/"
	}

	reqURL := baseURL + "/" + prefix + dbPath

	bootstrap.LogInfo("***** VAULT URL: %s", reqURL)

	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: vaultTransport(&tls.Config{Certificates: []tls.Certificate{cert}}),
	}

	data, err := requestWithRetry(client, reqURL, func(req *http.Request) {
		req.Header.Set("Content-Type", "application/json")
	})
	if err != nil {
		return Credentials{}, err
	}

	return decodeCyberArk(data)
}

func decodeCyberArk(data []byte) (Credentials, error) {
	var resp struct {
		ErrorCode string  `json:"ErrorCode"`
		Username  *string `json:"UserName"`
		Password  *string `json:"Content"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return Credentials{}, fmt.Errorf("decode CyberArk response: %w", err)
	}
	if resp.ErrorCode != "" {
		return Credentials{}, fmt.Errorf("error getting secrets from vault: %s", resp.ErrorCode)
	}
	if resp.Username == nil || *resp.Username == "" {
		return Credentials{}, fmt.Errorf("vault response from CyberArk does not contain a database username")
	}
	if resp.Password == nil {
		return Credentials{}, fmt.Errorf("vault response from CyberArk does not contain a database password")
	}

	return Credentials{Username: *resp.Username, Password: *resp.Password}, nil
}

func vaultTransport(tlsConfig *tls.Config) *http.Transport {
	transport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return &http.Transport{Proxy: http.ProxyFromEnvironment, TLSClientConfig: tlsConfig}
	}

	clone := transport.Clone()
	clone.TLSClientConfig = tlsConfig
	return clone
}

var sleep = time.Sleep

func requestWithRetry(client *http.Client, reqURL string, configure func(*http.Request)) ([]byte, error) {
	attempts := 3
	delay := 5 * time.Second

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		if attempt > 1 {
			bootstrap.LogInfo("**** Vault is not available. Waiting %s... ****", delay)
			sleep(delay)
		}

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, reqURL, nil)
		if err != nil {
			return nil, err
		}

		configure(req)
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		data, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, readErr
		}

		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			return nil, fmt.Errorf("vault request failed with status %s: %s", resp.Status, strings.TrimSpace(string(data)))
		}

		return data, nil
	}

	return nil, fmt.Errorf("vault is not available after %d attempts: %w", attempts, lastErr)
}
