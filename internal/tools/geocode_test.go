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

func TestNewGeocode_returns_tool_with_correct_name(t *testing.T) {
	c := client.New("http://localhost")
	tool, _ := NewGeocode(c)
	if tool.Name != "geocode" {
		t.Errorf("name = %q", tool.Name)
	}
}

func TestGeocode_handler_returns_results(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("q") != "Madrid" {
			t.Errorf("q = %q", q.Get("q"))
		}
		w.Write([]byte(`[{"name":"Madrid","lat":40.4,"lon":-3.7}]`))
	}))
	defer srv.Close()

	c := client.New(srv.URL)
	_, handler := NewGeocode(c)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"query": "Madrid",
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
	if !strings.Contains(text, "Madrid") {
		t.Errorf("text = %q", text)
	}
}

func TestGeocode_handler_with_lat_lon_bias(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("lat") == "" {
			t.Error("expected lat bias param")
		}
		if q.Get("lon") == "" {
			t.Error("expected lon bias param")
		}
		w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	c := client.New(srv.URL)
	_, handler := NewGeocode(c)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"query": "Calle Mayor",
				"lat":   40.4,
				"lon":   -3.7,
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
}

func TestGeocode_handler_without_lat_lon_bias(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("lat") != "" {
			t.Error("lat should be absent")
		}
		if q.Get("lon") != "" {
			t.Error("lon should be absent")
		}
		w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	c := client.New(srv.URL)
	_, handler := NewGeocode(c)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"query": "Barcelona",
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
}

func TestGeocode_handler_error_missing_query(t *testing.T) {
	c := client.New("http://localhost")
	_, handler := NewGeocode(c)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{},
		},
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected tool error for missing query")
	}
}

func TestGeocode_handler_error_on_api_failure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	c := client.New(srv.URL)
	_, handler := NewGeocode(c)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"query": "Madrid",
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

func TestGeocode_handler_with_only_lat_bias(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("lat") == "" {
			t.Error("expected lat")
		}
		// lon should not be set since it was not a float64
		w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	c := client.New(srv.URL)
	_, handler := NewGeocode(c)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"query": "Test",
				"lat":   40.4,
				// lon intentionally absent
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
}
