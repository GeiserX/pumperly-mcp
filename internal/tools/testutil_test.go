package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/geiserx/pumperly-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Fixtures in the shapes the live API returns (checked against pumperly.com).
const (
	nearestFixture = `{"type":"FeatureCollection","features":[
{"type":"Feature","geometry":{"type":"Point","coordinates":[-3.703139,40.405278]},"properties":{"id":"a","externalId":"4508","country":"ES","name":"Blanca Madrid","brand":"Blanca","address":"GLORIETA EMBAJADORES, 0","city":"Madrid","fuelType":"B7","currency":"EUR","price":1.879,"reportedAt":"2026-10-27T04:37:35.719Z","distanceKm":1.285}},
{"type":"Feature","geometry":{"type":"Point","coordinates":[-3.705389,40.404417]},"properties":{"id":"b","externalId":"3217","country":"ES","name":"Repsol Madrid","brand":"Repsol","address":"CL PASEO ACACIAS, 8","city":"Madrid","fuelType":"B7","currency":"EUR","price":1.995,"reportedAt":"2026-10-27T04:37:35.719Z","distanceKm":1.39}},
{"type":"Feature","geometry":{"type":"Point","coordinates":[-3.71,40.40]},"properties":{"id":"c","externalId":"9","country":"ES","name":"Ballenoil","brand":null,"address":"CL X, 1","city":"Madrid","fuelType":"B7","currency":"EUR","price":1.659,"reportedAt":"2026-10-27T04:37:35.719Z","distanceKm":2.5}},
{"type":"Feature","geometry":{"type":"Point","coordinates":[-3.72,40.41]},"properties":{"id":"d","externalId":"10","country":"ES","name":"No Price","brand":"X","address":"CL Y, 2","city":"Madrid","fuelType":"B7","currency":"EUR","distanceKm":0.5}}
]}`
	geocodeFixture = `[{"name":"Atocha","city":"Madrid","state":"Comunidad de Madrid","country":"España","coordinates":[-3.688582,40.406975]},{"name":"Atocha","city":"Madrid","state":null,"country":"España","coordinates":[-3.6886,40.4069]}]`
)

type recorded struct {
	Method string
	Path   string
	Query  map[string]string
	Body   map[string]any
}

// fakeAPI serves canned bodies per "METHOD /path" and records each request.
type fakeAPI struct {
	srv  *httptest.Server
	mu   sync.Mutex
	reqs []recorded
}

func newFakeAPI(t *testing.T, routes map[string]string) *fakeAPI {
	t.Helper()
	f := &fakeAPI{}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := recorded{Method: r.Method, Path: r.URL.Path, Query: map[string]string{}}
		for k := range r.URL.Query() {
			rec.Query[k] = r.URL.Query().Get(k)
		}
		if b, _ := io.ReadAll(r.Body); len(b) > 0 {
			_ = json.Unmarshal(b, &rec.Body)
		}
		f.mu.Lock()
		f.reqs = append(f.reqs, rec)
		f.mu.Unlock()
		body, ok := routes[r.Method+" "+r.URL.Path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":"Not found"}`))
			return
		}
		if len(body) > 0 && body[0] == '!' { // "!<status>" fails the request
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"Invalid parameters","details":[{"path":["fuel"],"message":"Invalid option"}]}`))
			return
		}
		w.Write([]byte(body))
	}))
	t.Cleanup(f.srv.Close)
	return f
}

func (f *fakeAPI) client() *client.Client { return client.New(f.srv.URL) }

func (f *fakeAPI) requests() []recorded {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]recorded(nil), f.reqs...)
}

func (f *fakeAPI) last(t *testing.T, path string) recorded {
	t.Helper()
	reqs := f.requests()
	for i := len(reqs) - 1; i >= 0; i-- {
		if reqs[i].Path == path {
			return reqs[i]
		}
	}
	t.Fatalf("no request to %s; got %+v", path, reqs)
	return recorded{}
}

func call(t *testing.T, h server.ToolHandlerFunc, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	res, err := h(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: args}})
	if err != nil {
		t.Fatalf("handler returned a Go error: %v", err)
	}
	return res
}

func text(t *testing.T, res *mcp.CallToolResult) string {
	t.Helper()
	if len(res.Content) == 0 {
		t.Fatal("no content")
	}
	tc, ok := res.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("content is %T", res.Content[0])
	}
	return tc.Text
}

func wantError(t *testing.T, res *mcp.CallToolResult, substr string) {
	t.Helper()
	if !res.IsError {
		t.Fatalf("want an error result containing %q, got %s", substr, text(t, res))
	}
	if got := text(t, res); !strings.Contains(got, substr) {
		t.Errorf("error = %q, want it to contain %q", got, substr)
	}
}

// as decodes the structured content into T, and checks the text content
// carries the same JSON.
func as[T any](t *testing.T, res *mcp.CallToolResult) T {
	t.Helper()
	if res.IsError {
		t.Fatalf("unexpected error result: %s", text(t, res))
	}
	var v T
	b, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatalf("structured content does not decode: %v", err)
	}
	if text(t, res) != string(b) {
		t.Errorf("text content differs from structured content:\n%s\n%s", text(t, res), b)
	}
	return v
}
