package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/geiserx/pumperly-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestNewFindRouteStations_returns_tool_with_correct_name(t *testing.T) {
	c := client.New("http://localhost")
	tool, _ := NewFindRouteStations(c)
	if tool.Name != "find_route_stations" {
		t.Errorf("name = %q", tool.Name)
	}
}

func TestFindRouteStations_handler_sends_geometry_and_fuel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var m map[string]any
		json.Unmarshal(body, &m)

		if _, ok := m["geometry"]; !ok {
			t.Error("missing geometry")
		}
		if m["fuel"] != "B7" {
			t.Errorf("fuel = %v", m["fuel"])
		}
		w.Write([]byte(`[{"station":"A"}]`))
	}))
	defer srv.Close()

	c := client.New(srv.URL)
	_, handler := NewFindRouteStations(c)

	geom := `{"type":"LineString","coordinates":[[-3.7,40.4],[-0.37,39.47]]}`
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"geometry": geom,
				"fuel":     "B7",
			},
		},
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		text := result.Content[0].(mcp.TextContent).Text
		t.Fatalf("handler returned error: %s", text)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "station") {
		t.Errorf("text = %q", text)
	}
}

func TestFindRouteStations_handler_uses_default_corridor(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var m map[string]any
		json.Unmarshal(body, &m)
		if corridor, ok := m["corridorKm"].(float64); !ok || corridor != 5 {
			t.Errorf("corridorKm = %v, want 5", m["corridorKm"])
		}
		w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	c := client.New(srv.URL)
	_, handler := NewFindRouteStations(c)

	geom := `{"type":"LineString","coordinates":[[-3.7,40.4],[-0.37,39.47]]}`
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"geometry": geom,
				"fuel":     "B7",
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

func TestFindRouteStations_handler_custom_corridor(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var m map[string]any
		json.Unmarshal(body, &m)
		if corridor, ok := m["corridorKm"].(float64); !ok || corridor != 15 {
			t.Errorf("corridorKm = %v, want 15", m["corridorKm"])
		}
		w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	c := client.New(srv.URL)
	_, handler := NewFindRouteStations(c)

	geom := `{"type":"LineString","coordinates":[[-3.7,40.4],[-0.37,39.47]]}`
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"geometry":    geom,
				"fuel":        "B7",
				"corridor_km": 15.0,
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

func TestFindRouteStations_handler_error_missing_geometry(t *testing.T) {
	c := client.New("http://localhost")
	_, handler := NewFindRouteStations(c)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"fuel": "B7",
			},
		},
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected tool error for missing geometry")
	}
}

func TestFindRouteStations_handler_error_missing_fuel(t *testing.T) {
	c := client.New("http://localhost")
	_, handler := NewFindRouteStations(c)

	geom := `{"type":"LineString","coordinates":[[-3.7,40.4]]}`
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"geometry": geom,
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

func TestFindRouteStations_handler_error_invalid_geometry_json(t *testing.T) {
	c := client.New("http://localhost")
	_, handler := NewFindRouteStations(c)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"geometry": "not-json",
				"fuel":     "B7",
			},
		},
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected tool error for invalid geometry JSON")
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "invalid geometry JSON") {
		t.Errorf("text = %q", text)
	}
}

func TestFindRouteStations_handler_error_on_api_failure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := client.New(srv.URL)
	_, handler := NewFindRouteStations(c)

	geom := `{"type":"LineString","coordinates":[[-3.7,40.4],[-0.37,39.47]]}`
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"geometry": geom,
				"fuel":     "B7",
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
