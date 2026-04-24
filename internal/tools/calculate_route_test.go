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

func TestNewCalculateRoute_returns_tool_with_correct_name(t *testing.T) {
	c := client.New("http://localhost")
	tool, _ := NewCalculateRoute(c)
	if tool.Name != "calculate_route" {
		t.Errorf("name = %q", tool.Name)
	}
}

func TestCalculateRoute_handler_sends_origin_and_destination(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var m map[string]any
		json.Unmarshal(body, &m)

		if _, ok := m["origin"]; !ok {
			t.Error("missing origin")
		}
		if _, ok := m["destination"]; !ok {
			t.Error("missing destination")
		}
		w.Write([]byte(`{"distance_km":350}`))
	}))
	defer srv.Close()

	c := client.New(srv.URL)
	_, handler := NewCalculateRoute(c)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"origin_lon": -3.7,
				"origin_lat": 40.4,
				"dest_lon":   -0.37,
				"dest_lat":   39.47,
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
	if !strings.Contains(text, "distance_km") {
		t.Errorf("text = %q", text)
	}
}

func TestCalculateRoute_handler_with_waypoints(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var m map[string]any
		json.Unmarshal(body, &m)
		if _, ok := m["waypoints"]; !ok {
			t.Error("missing waypoints")
		}
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := client.New(srv.URL)
	_, handler := NewCalculateRoute(c)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"origin_lon": -3.7,
				"origin_lat": 40.4,
				"dest_lon":   -0.37,
				"dest_lat":   39.47,
				"waypoints":  `[[-2.0, 39.0], [-1.0, 38.5]]`,
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

func TestCalculateRoute_handler_with_invalid_waypoints_json(t *testing.T) {
	c := client.New("http://localhost")
	_, handler := NewCalculateRoute(c)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"origin_lon": -3.7,
				"origin_lat": 40.4,
				"dest_lon":   -0.37,
				"dest_lat":   39.47,
				"waypoints":  "not-json",
			},
		},
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected tool error for invalid waypoints JSON")
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "invalid waypoints JSON") {
		t.Errorf("text = %q", text)
	}
}

func TestCalculateRoute_handler_with_empty_waypoints_string(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var m map[string]any
		json.Unmarshal(body, &m)
		if _, ok := m["waypoints"]; ok {
			t.Error("waypoints should not be set for empty string")
		}
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := client.New(srv.URL)
	_, handler := NewCalculateRoute(c)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"origin_lon": -3.7,
				"origin_lat": 40.4,
				"dest_lon":   -0.37,
				"dest_lat":   39.47,
				"waypoints":  "",
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

func TestCalculateRoute_handler_error_missing_origin_lon(t *testing.T) {
	c := client.New("http://localhost")
	_, handler := NewCalculateRoute(c)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"origin_lat": 40.4,
				"dest_lon":   -0.37,
				"dest_lat":   39.47,
			},
		},
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected tool error")
	}
}

func TestCalculateRoute_handler_error_missing_origin_lat(t *testing.T) {
	c := client.New("http://localhost")
	_, handler := NewCalculateRoute(c)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"origin_lon": -3.7,
				"dest_lon":   -0.37,
				"dest_lat":   39.47,
			},
		},
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected tool error")
	}
}

func TestCalculateRoute_handler_error_missing_dest_lon(t *testing.T) {
	c := client.New("http://localhost")
	_, handler := NewCalculateRoute(c)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"origin_lon": -3.7,
				"origin_lat": 40.4,
				"dest_lat":   39.47,
			},
		},
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected tool error")
	}
}

func TestCalculateRoute_handler_error_missing_dest_lat(t *testing.T) {
	c := client.New("http://localhost")
	_, handler := NewCalculateRoute(c)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"origin_lon": -3.7,
				"origin_lat": 40.4,
				"dest_lon":   -0.37,
			},
		},
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected tool error")
	}
}

func TestCalculateRoute_handler_error_on_api_failure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := client.New(srv.URL)
	_, handler := NewCalculateRoute(c)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"origin_lon": -3.7,
				"origin_lat": 40.4,
				"dest_lon":   -0.37,
				"dest_lat":   39.47,
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
