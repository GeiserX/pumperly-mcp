package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/geiserx/pumperly-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestNewGetStationsInArea_returns_tool_with_correct_name(t *testing.T) {
	c := client.New("http://localhost")
	tool, _ := NewGetStationsInArea(c)
	if tool.Name != "get_stations_in_area" {
		t.Errorf("name = %q", tool.Name)
	}
}

func TestGetStationsInArea_handler_returns_stations(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("bbox") != "-4,39,-3,41" {
			t.Errorf("bbox = %q", q.Get("bbox"))
		}
		if q.Get("fuel") != "E5" {
			t.Errorf("fuel = %q", q.Get("fuel"))
		}
		w.Write([]byte(`[{"id":1}]`))
	}))
	defer srv.Close()

	c := client.New(srv.URL)
	_, handler := NewGetStationsInArea(c)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"bbox": "-4,39,-3,41",
				"fuel": "E5",
			},
		},
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("handler returned error result")
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, `"id"`) {
		t.Errorf("text = %q", text)
	}
}

func TestGetStationsInArea_handler_error_missing_bbox(t *testing.T) {
	c := client.New("http://localhost")
	_, handler := NewGetStationsInArea(c)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"fuel": "E5",
			},
		},
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected tool error for missing bbox")
	}
}

func TestGetStationsInArea_handler_error_missing_fuel(t *testing.T) {
	c := client.New("http://localhost")
	_, handler := NewGetStationsInArea(c)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"bbox": "-4,39,-3,41",
			},
		},
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected tool error for missing fuel")
	}
}

func TestGetStationsInArea_handler_error_on_api_failure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()

	c := client.New(srv.URL)
	_, handler := NewGetStationsInArea(c)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"bbox": "-4,39,-3,41",
				"fuel": "E5",
			},
		},
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected tool error for API failure")
	}
}
