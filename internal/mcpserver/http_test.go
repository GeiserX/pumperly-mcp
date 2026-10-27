package mcpserver

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/geiserx/pumperly-mcp/client"
)

const modernToolsList = `{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{` + modernMeta + `}}`

func newTestHandler(t *testing.T, opts HTTPOptions) http.Handler {
	t.Helper()
	h, err := NewHTTPHandler(New(client.New("http://127.0.0.1:1")), opts)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func modernRequest(body io.Reader) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:8080/mcp", body)
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "application/json, text/event-stream")
	r.Header.Set("MCP-Protocol-Version", "2026-07-28")
	r.Header.Set("Mcp-Method", "tools/list")
	return r
}

func serve(h http.Handler, r *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestHTTP_modern_tools_list_is_stateless(t *testing.T) {
	w := serve(newTestHandler(t, HTTPOptions{}), modernRequest(strings.NewReader(modernToolsList)))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", w.Code, w.Body)
	}
	if id := w.Header().Get("Mcp-Session-Id"); id != "" {
		t.Errorf("stateless server minted a session id %q", id)
	}
	if !strings.Contains(w.Body.String(), `"resultType":"complete"`) {
		t.Errorf("body = %s", w.Body)
	}
}

func TestHTTP_legacy_initialize_still_works(t *testing.T) {
	body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"legacy","version":"0"}}}`
	r := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:8080/mcp", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "application/json, text/event-stream")
	w := serve(newTestHandler(t, HTTPOptions{}), r)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"protocolVersion":"2025-06-18"`) {
		t.Fatalf("status = %d, body %s", w.Code, w.Body)
	}
	if id := w.Header().Get("Mcp-Session-Id"); id != "" {
		t.Errorf("stateless server minted a session id %q for a legacy client", id)
	}
}

func TestHTTP_rejects_cross_origin_browser_requests(t *testing.T) {
	h := newTestHandler(t, HTTPOptions{})
	r := modernRequest(strings.NewReader(modernToolsList))
	r.Header.Set("Origin", "https://evil.example")
	r.Header.Set("Sec-Fetch-Site", "cross-site")
	if w := serve(h, r); w.Code != http.StatusForbidden {
		t.Errorf("cross-site status = %d, want 403", w.Code)
	}
	// Positive control: the same request from a non-browser client passes.
	if w := serve(h, modernRequest(strings.NewReader(modernToolsList))); w.Code != http.StatusOK {
		t.Errorf("control status = %d, want 200", w.Code)
	}
}

func TestHTTP_trusted_origin_is_allowed(t *testing.T) {
	h := newTestHandler(t, HTTPOptions{AllowedOrigins: []string{"https://app.example"}})
	r := modernRequest(strings.NewReader(modernToolsList))
	r.Header.Set("Origin", "https://app.example")
	r.Header.Set("Sec-Fetch-Site", "cross-site")
	if w := serve(h, r); w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestHTTP_rejects_oversized_bodies(t *testing.T) {
	h := newTestHandler(t, HTTPOptions{})
	big := bytes.Repeat([]byte("a"), MaxRequestBytes+1)

	if w := serve(h, modernRequest(bytes.NewReader(big))); w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("declared length: status = %d, want 413", w.Code)
	}
	// Chunked: no Content-Length, so only the reader cap can stop it.
	r := modernRequest(io.MultiReader(bytes.NewReader(big)))
	r.ContentLength = -1
	if w := serve(h, r); w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("chunked: status = %d, want 413", w.Code)
	}
}

func TestHTTP_bearer_token(t *testing.T) {
	h := newTestHandler(t, HTTPOptions{AuthToken: "tok-123"})
	for name, tc := range map[string]struct {
		header string
		want   int
	}{
		"missing":          {"", http.StatusUnauthorized},
		"wrong token":      {"Bearer tok-999", http.StatusUnauthorized},
		"no scheme":        {"tok-123", http.StatusUnauthorized},
		"scheme only":      {"Bearer", http.StatusUnauthorized},
		"basic scheme":     {"Basic tok-123", http.StatusUnauthorized},
		"right token":      {"Bearer tok-123", http.StatusOK},
		"lowercase scheme": {"bearer tok-123", http.StatusOK},
	} {
		r := modernRequest(strings.NewReader(modernToolsList))
		if tc.header != "" {
			r.Header.Set("Authorization", tc.header)
		}
		w := serve(h, r)
		if w.Code != tc.want {
			t.Errorf("%s: status = %d, want %d", name, w.Code, tc.want)
		}
		if w.Code == http.StatusUnauthorized && w.Header().Get("WWW-Authenticate") == "" {
			t.Errorf("%s: 401 without WWW-Authenticate", name)
		}
	}
}

func TestNewHTTPServer_requires_a_token_off_loopback(t *testing.T) {
	s := New(client.New("http://127.0.0.1:1"))
	if _, err := NewHTTPServer(s, HTTPOptions{Addr: "0.0.0.0:8080"}); err == nil {
		t.Error("0.0.0.0 without a token: want an error")
	}
	if _, err := NewHTTPServer(s, HTTPOptions{Addr: ":8080"}); err == nil {
		t.Error(":8080 without a token: want an error")
	}
	srv, err := NewHTTPServer(s, HTTPOptions{Addr: "0.0.0.0:8080", AuthToken: "tok"})
	if err != nil {
		t.Fatal(err)
	}
	if srv.ReadHeaderTimeout == 0 || srv.ReadTimeout == 0 || srv.IdleTimeout == 0 {
		t.Errorf("timeouts unset: %+v", srv)
	}
	if _, err := NewHTTPServer(s, HTTPOptions{Addr: "127.0.0.1:8080"}); err != nil {
		t.Errorf("loopback without a token: %v", err)
	}
}

func TestIsLoopback(t *testing.T) {
	for addr, want := range map[string]bool{
		"127.0.0.1:8080": true,
		"localhost:8080": true,
		"[::1]:8080":     true,
		"127.0.0.2:80":   true,
		"0.0.0.0:8080":   false,
		":8080":          false,
		"[::]:8080":      false,
		"192.168.1.5:80": false,
		"example.com:80": false,
	} {
		if got := IsLoopback(addr); got != want {
			t.Errorf("IsLoopback(%q) = %v, want %v", addr, got, want)
		}
	}
}

func TestHTTP_unknown_path_is_404(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:8080/other", strings.NewReader(modernToolsList))
	if w := serve(newTestHandler(t, HTTPOptions{}), r); w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}
