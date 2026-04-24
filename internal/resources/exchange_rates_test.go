package resources

import (
	"context"
	"strings"
	"testing"

	"github.com/geiserx/pumperly-mcp/client"
	"github.com/mark3labs/mcp-go/server"
)

func TestRegisterExchangeRates_returns_data(t *testing.T) {
	apiSrv := mockAPI(t, map[string]struct {
		status int
		body   string
	}{
		"GET /api/exchange-rates": {200, `{"EUR_USD":1.08}`},
	})
	c := client.New(apiSrv.URL)
	s := server.NewMCPServer("test", "0.0.1")
	RegisterExchangeRates(s, c)

	resp := s.HandleMessage(context.Background(), readResourceMsg(1, "pumperly://exchange-rates"))
	text := extractTextFromResponse(t, resp)
	if !strings.Contains(text, "EUR_USD") {
		t.Errorf("text = %q", text)
	}
}

func TestRegisterExchangeRates_returns_error_on_api_failure(t *testing.T) {
	apiSrv := mockAPI(t, map[string]struct {
		status int
		body   string
	}{
		"GET /api/exchange-rates": {503, "unavailable"},
	})
	c := client.New(apiSrv.URL)
	s := server.NewMCPServer("test", "0.0.1")
	RegisterExchangeRates(s, c)

	resp := s.HandleMessage(context.Background(), readResourceMsg(1, "pumperly://exchange-rates"))
	assertIsError(t, resp)
}
