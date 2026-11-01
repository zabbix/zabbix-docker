package httpapi

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

type responseBody struct {
	io.Reader
	closed bool
}

func (b *responseBody) Close() error { b.closed = true; return nil }

func TestDoBoundsAndClosesResponse(t *testing.T) {
	for _, size := range []int{1 << 20, (1 << 20) + 1} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			body := &responseBody{Reader: strings.NewReader(strings.Repeat("x", size))}
			client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: 500, Header: http.Header{"X-Test": {"kept"}}, Body: body}, nil
			})}
			req, err := http.NewRequest(http.MethodGet, "http://example.test", nil)
			if err != nil {
				t.Fatal(err)
			}
			resp, err := Do(client, req)
			if !body.closed {
				t.Error("response body not closed")
			}
			if size > 1<<20 {
				if err == nil || !strings.Contains(err.Error(), "exceeds 1 MiB") {
					t.Fatalf("oversized response: %v", err)
				}
			} else if err != nil || resp.StatusCode != 500 || resp.Header.Get("X-Test") != "kept" || len(resp.Body) != size {
				t.Fatalf("unexpected response: status=%d bytes=%d error=%v", resp.StatusCode, len(resp.Body), err)
			}
		})
	}
}

func TestDoPreservesErrorCause(t *testing.T) {
	cause := errors.New("connection refused")
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, cause
	})}
	req, err := http.NewRequest(http.MethodGet, "http://example.test/schema", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Do(client, req)
	if !errors.Is(err, cause) {
		t.Fatalf("original error is not preserved: %v", err)
	}
}

func TestErrorMessageRedactsBeforeTruncating(t *testing.T) {
	body := []byte("  " + strings.Repeat("x", 2040) + "private-secret  ")
	message := ErrorMessage(body, "private-secret")
	if len(message) != 2048 || strings.Contains(message, "private") || !strings.HasPrefix(message, "xxx") {
		t.Fatalf("unexpected diagnostic: length=%d suffix=%q", len(message), message[2040:])
	}
	if got := ErrorMessage([]byte("  failure  "), ""); got != "failure" {
		t.Fatalf("empty secret changed diagnostic: %q", got)
	}
}
