package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const requestTimeout = 2 * time.Second

func main() {
	url, err := healthcheckURL(os.Args[1:], os.Getenv)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	client := &http.Client{Timeout: requestTimeout}
	if err := checkHealth(client, url); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func healthcheckURL(args []string, getenv func(string) string) (string, error) {
	if len(args) == 1 && args[0] != "/ready" {
		return args[0], nil
	}
	if len(args) > 1 {
		return "", fmt.Errorf("usage: http-healthcheck [/ready|url]")
	}

	port := strings.TrimSpace(getenv("ZBX_LISTENHEALTHPORT"))
	if port == "" {
		port = "8081"
	}

	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return "", fmt.Errorf("invalid ZBX_LISTENHEALTHPORT %q: expected a port between 1 and 65535", port)
	}

	path := "/health"
	if len(args) == 1 {
		path = args[0]
	}
	return fmt.Sprintf("http://127.0.0.1:%d%s", portNumber, path), nil
}

func checkHealth(client *http.Client, targetURL string) error {
	endpoint, err := url.Parse(targetURL)
	if err != nil {
		return err
	}
	response, err := client.Get(targetURL) //nolint:noctx // The client timeout bounds the health check.
	if err != nil {
		return err
	}

	body, copyErr := io.ReadAll(response.Body)
	closeErr := response.Body.Close()

	if copyErr != nil {
		return copyErr
	}

	if closeErr != nil {
		return closeErr
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("unexpected HTTP status: %s", response.Status)
	}

	if endpoint.Path == "/ready" {
		var result struct {
			Ready bool `json:"ready"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			return fmt.Errorf("invalid readiness response: %w", err)
		}
		if !result.Ready {
			return fmt.Errorf("Bridge Adapter is not ready")
		}
	}
	return nil
}
