package elasticsearch

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/httpapi"
)

// Client prepares history storage through the Elasticsearch HTTP API.
type Client struct {
	http               *http.Client
	endpoint           string
	query              url.Values
	username, password string
}

// NewClient extracts URL credentials for HTTP Basic authentication.
func NewClient(endpoint string) (*Client, error) {
	u, err := url.Parse(endpoint)
	if err != nil || u.Hostname() == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return nil, fmt.Errorf("Elasticsearch URL %q must be an http(s) endpoint", endpoint)
	}

	c := &Client{http: httpapi.NewClient(), query: u.Query()}

	if u.User != nil {
		c.username = u.User.Username()
		c.password, _ = u.User.Password()
		u.User = nil
	}

	u.RawQuery, u.Fragment = "", ""
	u.ForceQuery = false
	c.endpoint = strings.TrimRight(u.String(), "/")

	return c, nil
}

// Endpoint returns the URL without credentials.
func (c *Client) Endpoint() string { return c.endpoint }

func (c *Client) Close() { c.http.CloseIdleConnections() }

type responseError struct {
	status  int
	message string
}

func (e *responseError) Error() string { return e.message }

func (c *Client) request(ctx context.Context, method, path string, body []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, c.endpoint+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	for key, values := range c.query {
		if !query.Has(key) {
			query[key] = values
		}
	}
	req.URL.RawQuery = query.Encode()

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.username != "" || c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	resp, err := httpapi.Do(c.http, req)
	if err != nil {
		return nil, fmt.Errorf("Elasticsearch %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		message := httpapi.ErrorMessage(resp.Body, c.password)
		return nil, &responseError{resp.StatusCode, fmt.Sprintf("Elasticsearch HTTP %d: %s", resp.StatusCode, message)}
	}

	return resp.Body, nil
}

func (c *Client) exists(ctx context.Context, method, path string) (bool, error) {
	_, err := c.request(ctx, method, path, nil)

	var response *responseError
	if errors.As(err, &response) && response.status == http.StatusNotFound {
		return false, nil
	}

	return err == nil, err
}
