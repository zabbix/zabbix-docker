package clickhouse

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/httpapi"
)

const (
	queryTimeout          = 60 * time.Second
	distributedDDLTimeout = 180 * time.Second
	// Allow ClickHouse to finish waiting for cluster DDL and return its result.
	clusterRequestTimeout = distributedDDLTimeout + 30*time.Second
)

type Connection struct {
	URL, Username, Password string
	SourceIP                string
	CADirectory             string
	CertFile, KeyFile       string
	KeyPassword             string
	VerifyPeer, VerifyHost  bool
}

// Client executes SQL through the ClickHouse HTTP interface.
type Client struct {
	http            *http.Client
	endpoint        string
	clusterEndpoint string
	connection      Connection
}

// NewClient validates the connection and configures its HTTP transport.
func NewClient(ctx context.Context, conn Connection) (*Client, error) {
	u, err := url.Parse(conn.URL)
	if err != nil || u.Hostname() == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return nil, fmt.Errorf("ClickHouse URL %q must be an http(s) endpoint", conn.URL)
	}

	client := httpapi.NewClient()
	transport := client.Transport.(*http.Transport)

	if conn.SourceIP != "" {
		ip := net.ParseIP(conn.SourceIP)
		if ip == nil {
			return nil, fmt.Errorf("invalid ClickHouse source_ip %q", conn.SourceIP)
		}
		dialer := &net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second, LocalAddr: &net.TCPAddr{IP: ip}}
		transport.DialContext = dialer.DialContext
	}

	if u.Scheme == "https" {
		transport.TLSClientConfig, err = conn.tlsConfig(ctx, u.Hostname())
		if err != nil {
			return nil, err
		}
	}

	q := u.Query()
	// Buffer results to reduce errors arriving after HTTP headers are sent.
	// Progress headers can still cause an error response to have HTTP status 200.
	q.Set("wait_end_of_query", "1")
	// Check every replica even if the user profile permits unavailable shards.
	q.Set("skip_unavailable_shards", "0")
	u.RawQuery = q.Encode()
	endpoint := u.String()
	q.Set("distributed_ddl_task_timeout", fmt.Sprint(int(distributedDDLTimeout/time.Second)))
	u.RawQuery = q.Encode()

	return &Client{
		connection: conn, endpoint: endpoint, clusterEndpoint: u.String(),
		http: client,
	}, nil
}

func (c *Client) Close() { c.http.CloseIdleConnections() }

func (c *Client) query(ctx context.Context, sql string) ([]byte, error) {
	return c.request(ctx, sql, false)
}

func (c *Client) request(ctx context.Context, sql string, clusterDDL bool) ([]byte, error) {
	timeout := queryTimeout
	endpoint := c.endpoint

	if clusterDDL {
		timeout = clusterRequestTimeout
		endpoint = c.clusterEndpoint
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(sql))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	if c.connection.Username != "" || c.connection.Password != "" {
		req.SetBasicAuth(c.connection.Username, c.connection.Password)
	}

	resp, err := httpapi.Do(c.http, req)
	if err != nil {
		return nil, fmt.Errorf("ClickHouse %w", err)
	}

	if resp.StatusCode != http.StatusOK || resp.Header.Get("X-ClickHouse-Exception-Code") != "" ||
		strings.HasPrefix(strings.TrimSpace(string(resp.Body)), "Code:") {
		message := httpapi.ErrorMessage(resp.Body, c.connection.Password)
		return nil, fmt.Errorf("ClickHouse HTTP %d: %s", resp.StatusCode, message)
	}

	return resp.Body, nil
}

func (c Connection) tlsConfig(ctx context.Context, hostname string) (*tls.Config, error) {
	config := &tls.Config{ServerName: hostname}

	if c.VerifyPeer {
		roots, err := x509.SystemCertPool()
		if err != nil {
			return nil, fmt.Errorf("load system CA certificates: %w", err)
		}

		if c.CADirectory != "" {
			entries, err := os.ReadDir(c.CADirectory)
			if err != nil {
				return nil, fmt.Errorf("read ClickHouse CA directory: %w", err)
			}

			for _, entry := range entries {
				path := filepath.Join(c.CADirectory, entry.Name())
				info, err := os.Stat(path)
				if err != nil {
					return nil, fmt.Errorf("stat ClickHouse CA entry %q: %w", path, err)
				}
				if info.IsDir() {
					continue
				}

				data, err := os.ReadFile(path)
				if err != nil {
					return nil, fmt.Errorf("read ClickHouse CA certificate: %w", err)
				}

				if !roots.AppendCertsFromPEM(data) {
					bootstrap.LogWarn("** No PEM certificates parsed from ClickHouse CA file %q", path)
				}
			}
		}

		config.RootCAs = roots
	}

	if !c.VerifyPeer || !c.VerifyHost {
		config.InsecureSkipVerify = true

		config.VerifyConnection = func(state tls.ConnectionState) error {
			if c.VerifyPeer {
				if err := bootstrap.VerifyCertificateChain(state.PeerCertificates, config.RootCAs); err != nil {
					return err
				}
			}

			if c.VerifyHost {
				return state.PeerCertificates[0].VerifyHostname(hostname)
			}

			return nil
		}
	}

	if (c.CertFile == "") != (c.KeyFile == "") {
		return nil, errors.New("ssl_cert_file and ssl_key_file must be set together")
	}

	if c.CertFile != "" {
		certPEM, err := os.ReadFile(c.CertFile)
		if err != nil {
			return nil, fmt.Errorf("read ClickHouse client certificate: %w", err)
		}

		keyPEM, err := readPrivateKey(ctx, c.KeyFile, c.KeyPassword)
		if err != nil {
			return nil, fmt.Errorf("read ClickHouse client key: %w", err)
		}

		cert, err := tls.X509KeyPair(certPEM, keyPEM)
		if err != nil {
			return nil, fmt.Errorf("load ClickHouse client certificate: %w", err)
		}

		config.Certificates = []tls.Certificate{cert}
	}

	return config, nil
}

func readPrivateKey(ctx context.Context, path, password string) ([]byte, error) {
	if password == "" {
		return os.ReadFile(path)
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	command := exec.CommandContext(ctx, "openssl", "pkey", "-in", path, "-passin", "stdin")
	command.Stdin = strings.NewReader(password + "\n")
	var stderr bytes.Buffer
	command.Stderr = &stderr

	data, err := command.Output()
	if ctx.Err() != nil {
		return nil, fmt.Errorf("openssl pkey: %w", ctx.Err())
	}

	if err != nil {
		if message := strings.TrimSpace(stderr.String()); message != "" {
			return nil, fmt.Errorf("openssl pkey: %w: %s", err, message)
		}

		return nil, fmt.Errorf("openssl pkey: %w", err)
	}

	return data, nil
}
