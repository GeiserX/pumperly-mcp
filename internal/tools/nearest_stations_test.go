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

func TestNewFindNearestStations_returns_tool_with_correct_name(t *testing.T) {
	c := client.New("http://localhost")
	tool, _ := NewFindNearestStations(c)
	if tool.Name != "find_nearest_stations" {
		t.Errorf("name = %q", tool.Name)
	}
}

func TestFindNearestStations_handler_returns_stations(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("fuel") != "B7" {
			t.Errorf("fuel = %q", q.Get("fuel"))
		}
		w.Write([]byte(`[{"name":"Station A"}]`))
	}))
	defer srv.Close()

	c := client.New(srv.URL)
	_, handler := NewFindNearestStations(c)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"lat":  40.4,
				"lon":  -3.7,
				"fuel": "B7",
			},
		},
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("handler returned error result")
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "Station A") {
		t.Errorf("text = %q", text)
	}
}

func TestFindNearestStations_handler_uses_default_radius_and_limit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("radius_km") != "10" {
			t.Errorf("radius_km = %q, want 10", q.Get("radius_km"))
		}
		if q.Get("limit") != "5" {
			t.Errorf("limit = %q, want 5", q.Get("limit"))
		}
		w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	c := client.New(srv.URL)
	_, handler := NewFindNearestStations(c)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"lat":  40.4,
				"lon":  -3.7,
				"fuel": "B7",
			},
		},
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("handler returned error result")
	}
}

func TestFindNearestStations_handler_custom_radius_and_limit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("radius_km") != "25" {
			t.Errorf("radius_km = %q, want 25", q.Get("radius_km"))
		}
		if q.Get("limit") != "20" {
			t.Errorf("limit = %q, want 20", q.Get("limit"))
		}
		w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	c := client.New(srv.URL)
	_, handler := NewFindNearestStations(c)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"lat":       40.4,
				"lon":       -3.7,
				"fuel":      "B7",
				"radius_km": 25.0,
				"limit":     20.0,
			},
		},
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("handler returned error result")
	}
}

func TestFindNearestStations_handler_error_missing_lat(t *testing.T) {
	c := client.New("http://localhost")
	_, handler := NewFindNearestStations(c)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"lon":  -3.7,
				"fuel": "B7",
			},
		},
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected tool error for missing lat")
	}
}

func TestFindNearestStations_handler_error_missing_lon(t *testing.T) {
	c := client.New("http://localhost")
	_, handler := NewFindNearestStations(c)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"lat":  40.4,
				"fuel": "B7",
			},
		},
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected tool error for missing lon")
	}
}

func TestFindNearestStations_handler_error_missing_fuel(t *testing.T) {
	c := client.New("http://localhost")
	_, handler := NewFindNearestStations(c)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"lat": 40.4,
				"lon": -3.7,
			},
		},
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected tool error for missing fuel")
	}
}

func TestFindNearestStations_handler_error_on_api_failure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("boom"))
	}))
	defer srv.Close()

	c := client.New(srv.URL)
	_, handler := NewFindNearestStations(c)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"lat":  40.4,
				"lon":  -3.7,
				"fuel": "B7",
			},
		},
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected tool error for API failure")
	}
}
