package resources

import (
	"context"
	"strings"
	"testing"

	"github.com/geiserx/pumperly-mcp/client"
	"github.com/mark3labs/mcp-go/server"
)

func TestRegisterStats_returns_data(t *testing.T) {
	apiSrv := mockAPI(t, map[string]struct {
		status int
		body   string
	}{
		"GET /api/stats": {200, `{"stations":1234}`},
	})
	c := client.New(apiSrv.URL)
	s := server.NewMCPServer("test", "0.0.1")
	RegisterStats(s, c)

	resp := s.HandleMessage(context.Background(), readResourceMsg(1, "pumperly://stats"))
	text := extractTextFromResponse(t, resp)
	if !strings.Contains(text, "stations") {
		t.Errorf("text = %q", text)
	}
}

func TestRegisterStats_returns_error_on_api_failure(t *testing.T) {
	apiSrv := mockAPI(t, map[string]struct {
		status int
		body   string
	}{
		"GET /api/stats": {502, "bad gateway"},
	})
	c := client.New(apiSrv.URL)
	s := server.NewMCPServer("test", "0.0.1")
	RegisterStats(s, c)

	resp := s.HandleMessage(context.Background(), readResourceMsg(1, "pumperly://stats"))
	assertIsError(t, resp)
}
