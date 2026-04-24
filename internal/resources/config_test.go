package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/geiserx/pumperly-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func mockAPI(t *testing.T, routes map[string]struct {
	status int
	body   string
}) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Method + " " + r.URL.Path
		if route, ok := routes[key]; ok {
			w.WriteHeader(route.status)
			w.Write([]byte(route.body))
			return
		}
		w.WriteHeader(404)
		w.Write([]byte("not found: " + key))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func readResourceMsg(id int, uri string) []byte {
	return []byte(fmt.Sprintf(`{
		"jsonrpc": "2.0",
		"id": %d,
		"method": "resources/read",
		"params": {"uri": %q}
	}`, id, uri))
}

func extractTextFromResponse(t *testing.T, resp mcp.JSONRPCMessage) string {
	t.Helper()
	jsonResp, ok := resp.(mcp.JSONRPCResponse)
	if !ok {
		t.Fatalf("expected JSONRPCResponse, got %T: %+v", resp, resp)
	}
	b, err := json.Marshal(jsonResp.Result)
	if err != nil {
		t.Fatalf("marshal result: %v", err)
	}
	var result struct {
		Contents []struct {
			URI      string `json:"uri"`
			MIMEType string `json:"mimeType"`
			Text     string `json:"text"`
		} `json:"contents"`
	}
	if err := json.Unmarshal(b, &result); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if len(result.Contents) == 0 {
		t.Fatal("no contents in response")
	}
	return result.Contents[0].Text
}

func assertIsError(t *testing.T, resp mcp.JSONRPCMessage) {
	t.Helper()
	if _, ok := resp.(mcp.JSONRPCError); !ok {
		t.Fatalf("expected JSONRPCError, got %T: %+v", resp, resp)
	}
}

func TestRegisterConfig_returns_data(t *testing.T) {
	apiSrv := mockAPI(t, map[string]struct {
		status int
		body   string
	}{
		"GET /api/config": {200, `{"countries":["ES","PT"]}`},
	})
	c := client.New(apiSrv.URL)
	s := server.NewMCPServer("test", "0.0.1")
	RegisterConfig(s, c)

	resp := s.HandleMessage(context.Background(), readResourceMsg(1, "pumperly://config"))
	text := extractTextFromResponse(t, resp)
	if !strings.Contains(text, "countries") {
		t.Errorf("text = %q", text)
	}
}

func TestRegisterConfig_returns_error_on_api_failure(t *testing.T) {
	apiSrv := mockAPI(t, map[string]struct {
		status int
		body   string
	}{
		"GET /api/config": {500, "boom"},
	})
	c := client.New(apiSrv.URL)
	s := server.NewMCPServer("test", "0.0.1")
	RegisterConfig(s, c)

	resp := s.HandleMessage(context.Background(), readResourceMsg(1, "pumperly://config"))
	assertIsError(t, resp)
}

func TestRegisterConfig_response_has_correct_uri_and_mime(t *testing.T) {
	apiSrv := mockAPI(t, map[string]struct {
		status int
		body   string
	}{
		"GET /api/config": {200, `{}`},
	})
	c := client.New(apiSrv.URL)
	s := server.NewMCPServer("test", "0.0.1")
	RegisterConfig(s, c)

	resp := s.HandleMessage(context.Background(), readResourceMsg(1, "pumperly://config"))
	jsonResp, ok := resp.(mcp.JSONRPCResponse)
	if !ok {
		t.Fatalf("expected JSONRPCResponse, got %T", resp)
	}
	b, _ := json.Marshal(jsonResp.Result)
	resultStr := string(b)
	if !strings.Contains(resultStr, "pumperly://config") {
		t.Errorf("missing URI in response: %s", resultStr)
	}
	if !strings.Contains(resultStr, "application/json") {
		t.Errorf("missing MIME type in response: %s", resultStr)
	}
}
