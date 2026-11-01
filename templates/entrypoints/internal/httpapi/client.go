// Package httpapi provides bounded HTTP requests for schema preparation.
package httpapi

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Response struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

// NewClient creates an HTTP client that does not follow redirects.
func NewClient() *http.Client {
	return &http.Client{
		Transport: http.DefaultTransport.(*http.Transport).Clone(),
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

// Do limits response bodies to 1 MiB and leaves HTTP status handling to the caller.
func Do(client *http.Client, req *http.Request) (Response, error) {
	resp, err := client.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()
	const maxResponse = 1 << 20
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxResponse+1))
	if err != nil {
		return Response{}, err
	}
	if len(data) > maxResponse {
		return Response{}, errors.New("response exceeds 1 MiB")
	}
	return Response{StatusCode: resp.StatusCode, Header: resp.Header, Body: data}, nil
}

// ErrorMessage redacts the password and limits the response text included in errors.
func ErrorMessage(body []byte, secret string) string {
	message := strings.TrimSpace(string(body))
	if secret != "" {
		message = strings.ReplaceAll(message, secret, "[redacted]")
	}
	if len(message) > 2048 {
		message = message[:2048]
	}
	return message
}
