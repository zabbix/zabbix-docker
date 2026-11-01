package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(client.Close)
	return client
}

func TestPrepareRechecksAfterCreateError(t *testing.T) {
	for _, test := range []struct {
		name   string
		dated  bool
		status int
	}{
		{"index created concurrently", false, http.StatusOK},
		{"template created concurrently", true, http.StatusOK},
		{"creation failed", false, http.StatusNotFound},
		{"recheck failed", false, http.StatusServiceUnavailable},
	} {
		t.Run(test.name, func(t *testing.T) {
			path, errorKind := "/uint", "resource_already_exists_exception"
			if test.dated {
				path, errorKind = "/_index_template/uint_template", "illegal_argument_exception"
			}
			put := false
			client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/_ingest/pipeline/uint-pipeline":
				case "/_template/uint_template":
					w.WriteHeader(http.StatusNotFound)
				case path:
					if r.Method == http.MethodPut {
						put = true
						if test.dated && r.URL.Query().Get("create") != "true" {
							t.Error("template creation must use create=true")
						}
						w.WriteHeader(http.StatusBadRequest)
						fmt.Fprintf(w, `{"error":{"type":%q}}`, errorKind)
					} else if put {
						w.WriteHeader(test.status)
					} else {
						w.WriteHeader(http.StatusNotFound)
					}
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			})
			err := client.Prepare(context.Background(), Schema{ValueTypes: []string{"uint"}, DateIndex: test.dated, Shards: 1}, true)
			if !put {
				t.Fatal("creation was not attempted")
			}
			if test.status == http.StatusOK {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), errorKind) {
				t.Fatalf("expected creation error, got %v", err)
			}
			if test.status == http.StatusServiceUnavailable && !strings.Contains(err.Error(), "HTTP 503") {
				t.Fatalf("expected recheck error, got %v", err)
			}
		})
	}
}

func TestCreatedObjectVerification(t *testing.T) {
	for _, test := range []struct {
		name, body, wantErr string
		status              int
	}{
		{"not acknowledged but present", `{"acknowledged":false}`, "", http.StatusOK},
		{"acknowledged but missing", `{"acknowledged":true}`, "still missing", http.StatusNotFound},
		{"verification failed", `{}`, "HTTP 503", http.StatusServiceUnavailable},
	} {
		t.Run(test.name, func(t *testing.T) {
			put := false
			client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPut {
					put = true
					fmt.Fprint(w, test.body)
					return
				}
				if !put {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				w.WriteHeader(test.status)
			})
			err := client.Prepare(context.Background(), Schema{ValueTypes: []string{"uint"}, Shards: 1}, true)
			if test.wantErr == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("error = %v, want %q", err, test.wantErr)
			}
		})
	}
}

func TestPrepareExistingObjectsDoesNotWrite(t *testing.T) {
	for _, dated := range []bool{false, true} {
		for _, create := range []bool{false, true} {
			t.Run(fmt.Sprintf("date=%t/create=%t", dated, create), func(t *testing.T) {
				calls := 0
				client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
					calls++
					if r.Method != http.MethodHead && r.Method != http.MethodGet {
						t.Errorf("unexpected mutation: %s %s", r.Method, r.URL.Path)
					}
					if strings.Contains(r.URL.Path, "uint") {
						t.Error("queried a history type not included in types")
					}
					fmt.Fprint(w, `{}`)
				})
				if err := client.Prepare(context.Background(), Schema{ValueTypes: []string{"str", "json"}, DateIndex: dated, Shards: 5, Replicas: 1}, create); err != nil {
					t.Fatal(err)
				}
				want := 3
				if dated {
					want = 4
				}
				if calls != want {
					t.Fatalf("requests=%d, want %d", calls, want)
				}
			})
		}
	}
}

func TestPrepareCheckFailures(t *testing.T) {
	for _, status := range []int{404, 401, 403, 503} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			calls := 0
			client := testClient(t, func(w http.ResponseWriter, r *http.Request) { calls++; w.WriteHeader(status) })
			err := client.Prepare(context.Background(), Schema{ValueTypes: []string{"uint"}, Shards: 1}, false)
			if err == nil || calls != 1 {
				t.Fatalf("error=%v, requests=%d", err, calls)
			}
			if status == 404 && !strings.Contains(err.Error(), "creation is disabled") {
				t.Fatal(err)
			}
		})
	}
}

func TestLegacyTemplateIsPreserved(t *testing.T) {
	var paths []string
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.Method+" "+r.URL.Path)
		if r.Method == http.MethodPut {
			t.Error("existing legacy template must not be replaced")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/_index_template/") {
			w.WriteHeader(404)
		}
	})
	if err := client.Prepare(context.Background(), Schema{ValueTypes: []string{"str"}, DateIndex: true, Shards: 1}, true); err != nil {
		t.Fatal(err)
	}
	if len(paths) != 3 || paths[2] != "HEAD /_template/str_template" {
		t.Fatal(paths)
	}
}

func TestInvalidURLDiagnostic(t *testing.T) {
	for _, address := range []string{"invalid", "ftp://elasticsearch:9200", "http://elasticsearch:bad"} {
		_, err := NewClient(address)
		if err == nil || !strings.Contains(err.Error(), fmt.Sprintf("%q", address)) {
			t.Fatalf("unexpected URL error: %v", err)
		}
	}
}

func TestClientQueryParameters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		create := "false"
		if r.Method == http.MethodPut {
			create = "true"
		}
		if r.URL.Path != "/prefix/_index_template/uint_template" ||
			r.URL.Query().Get("pretty") != "true" || r.URL.Query().Get("create") != create {
			t.Errorf("unexpected request URL: %s", r.URL)
		}
	}))
	defer server.Close()
	client, err := NewClient(server.URL + "/prefix/?pretty=true&create=false#section")
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	for _, request := range []struct{ method, path string }{
		{http.MethodPut, "/_index_template/uint_template?create=true"},
		{http.MethodGet, "/_index_template/uint_template"},
	} {
		if _, err := client.request(context.Background(), request.method, request.path, nil); err != nil {
			t.Fatal(err)
		}
	}
}

func TestClientCredentialsAndRedirects(t *testing.T) {
	endpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, password, ok := r.BasicAuth()
		if !ok || user != "user" || password != "private-password" {
			t.Error("credentials not forwarded")
		}
		if r.URL.Path != "/prefix/uint" {
			t.Error(r.URL.Path)
		}
		w.Header().Set("Location", "http://127.0.0.1:1/leak")
		w.WriteHeader(http.StatusTemporaryRedirect)
		fmt.Fprint(w, "private-password")
	}))
	defer endpoint.Close()
	client, err := NewClient(strings.Replace(endpoint.URL, "://", "://user:private-password@", 1) + "/prefix/")
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	if strings.Contains(client.Endpoint(), "user") || strings.Contains(client.Endpoint(), "password") {
		t.Fatal("endpoint leaks credentials")
	}
	_, err = client.request(context.Background(), http.MethodGet, "/uint", nil)
	if err == nil || !strings.Contains(err.Error(), "307") || strings.Contains(err.Error(), "private-password") {
		t.Fatalf("error=%v", err)
	}
}

func TestCreateIndicesAndJSONPipeline(t *testing.T) {
	created := make(map[string]map[string]any)
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			if _, exists := created[r.URL.Path]; exists {
				t.Errorf("existing object overwritten: %s", r.URL.Path)
			}
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			created[r.URL.Path] = body
			fmt.Fprint(w, `{"acknowledged":true}`)
			return
		}
		if _, exists := created[r.URL.Path]; !exists {
			w.WriteHeader(http.StatusNotFound)
		}
	})
	schema := Schema{ValueTypes: []string{"uint", "json", "uint"}, Shards: 2, Replicas: 0}
	for _, create := range []bool{true, true, false} {
		if err := client.Prepare(context.Background(), schema, create); err != nil {
			t.Fatal(err)
		}
	}
	if len(created) != 3 || created["/uint"] == nil || created["/json"] == nil || created["/_ingest/pipeline/json"] == nil {
		t.Fatalf("unexpected created objects: %v", created)
	}
	for _, path := range []string{"/uint", "/json"} {
		settings := created[path]["settings"].(map[string]any)
		if settings["number_of_shards"] != float64(2) || settings["number_of_replicas"] != float64(0) {
			t.Errorf("%s settings: %v", path, settings)
		}
		pipeline, assigned := settings["default_pipeline"]
		if (path == "/json" && pipeline != "json") || (path == "/uint" && assigned) {
			t.Errorf("%s pipeline: %v", path, pipeline)
		}
	}
}
