package main

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHealthcheckURL(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name    string
		args    []string
		port    string
		want    string
		wantErr bool
	}{
		{name: "default", want: "http://127.0.0.1:8081/health"},
		{name: "custom port", port: "9091", want: "http://127.0.0.1:9091/health"},
		{name: "whitespace", port: " 9091 ", want: "http://127.0.0.1:9091/health"},
		{name: "empty after trimming", port: " ", want: "http://127.0.0.1:8081/health"},
		{name: "readiness default port", args: []string{"/ready"}, want: "http://127.0.0.1:8081/ready"},
		{name: "readiness custom port", args: []string{"/ready"}, port: "9091", want: "http://127.0.0.1:9091/ready"},
		{name: "readiness invalid port", args: []string{"/ready"}, port: "invalid", wantErr: true},
		{name: "explicit URL", args: []string{"http://localhost:9000/ready"}, port: "invalid", want: "http://localhost:9000/ready"},
		{name: "invalid port", port: "invalid", wantErr: true},
		{name: "zero port", port: "0", wantErr: true},
		{name: "negative port", port: "-1", wantErr: true},
		{name: "port out of range", port: "65536", wantErr: true},
		{name: "URL injection", port: "8081/ready", wantErr: true},
		{name: "too many arguments", args: []string{"one", "two"}, wantErr: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			url, err := healthcheckURL(test.args, func(name string) string {
				if name != "ZBX_LISTENHEALTHPORT" {
					t.Errorf("unexpected environment variable: %s", name)
				}
				return test.port
			})
			if (err != nil) != test.wantErr {
				t.Fatalf("healthcheckURL() error = %v, wantErr = %v", err, test.wantErr)
			}
			if url != test.want {
				t.Errorf("healthcheckURL() = %q, want %q", url, test.want)
			}
		})
	}
}

func TestCheckHealth(t *testing.T) {
	t.Parallel()

	for _, status := range []int{http.StatusOK, http.StatusNoContent, http.StatusNotFound, http.StatusServiceUnavailable} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != "/health" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
				w.WriteHeader(status)
			}))
			defer server.Close()

			_, port, err := net.SplitHostPort(server.Listener.Addr().String())
			if err != nil {
				t.Fatal(err)
			}
			url, err := healthcheckURL(nil, func(string) string { return port })
			if err != nil {
				t.Fatal(err)
			}
			err = checkHealth(server.Client(), url)
			wantErr := status >= http.StatusMultipleChoices
			if (err != nil) != wantErr {
				t.Fatalf("checkHealth() error = %v, wantErr = %v", err, wantErr)
			}
			if wantErr && !strings.Contains(err.Error(), "unexpected HTTP status") {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestCheckHealthTimeout(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	client := server.Client()
	client.Timeout = 20 * time.Millisecond
	if err := checkHealth(client, server.URL+"/health"); err == nil {
		t.Fatal("expected a timeout error")
	}
}

func TestCheckHealthConnectionFailure(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	server.Close()

	client := &http.Client{Timeout: requestTimeout}
	if err := checkHealth(client, server.URL+"/health"); err == nil {
		t.Fatal("expected a connection error")
	}
}

func TestCheckReadiness(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name    string
		body    string
		wantErr bool
	}{
		{name: "ready", body: `{"ready":true,"components":{"stream_opened":true}}`},
		{name: "not ready despite HTTP 200", body: `{"ready":false}`, wantErr: true},
		{name: "missing ready field", body: `{}`, wantErr: true},
		{name: "wrong ready type", body: `{"ready":"true"}`, wantErr: true},
		{name: "invalid JSON", body: `not JSON`, wantErr: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/ready" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				fmt.Fprint(w, test.body)
			}))
			defer server.Close()
			err := checkHealth(server.Client(), server.URL+"/ready")
			if (err != nil) != test.wantErr {
				t.Fatalf("checkHealth() error = %v, wantErr = %v", err, test.wantErr)
			}
		})
	}
}
