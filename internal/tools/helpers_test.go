package tools

import (
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// ── toFloat64 ────────────────────────────────────────────────────────────────

func TestToFloat64(t *testing.T) {
	tests := []struct {
		name   string
		input  any
		wantF  float64
		wantOK bool
	}{
		{name: "float64", input: float64(3.14), wantF: 3.14, wantOK: true},
		{name: "float32", input: float32(2.5), wantF: 2.5, wantOK: true},
		{name: "int", input: int(42), wantF: 42, wantOK: true},
		{name: "int64", input: int64(100), wantF: 100, wantOK: true},
		{name: "json.Number valid", input: json.Number("9.81"), wantF: 9.81, wantOK: true},
		{name: "json.Number invalid", input: json.Number("notnum"), wantF: 0, wantOK: false},
		{name: "string valid", input: "123.456", wantF: 123.456, wantOK: true},
		{name: "string invalid", input: "abc", wantF: 0, wantOK: false},
		{name: "bool unsupported", input: true, wantF: 0, wantOK: false},
		{name: "nil unsupported", input: nil, wantF: 0, wantOK: false},
		{name: "slice unsupported", input: []int{1}, wantF: 0, wantOK: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f, ok := toFloat64(tc.input)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
			if ok && f != tc.wantF {
				t.Errorf("value = %v, want %v", f, tc.wantF)
			}
		})
	}
}

// ── requiredFloat ────────────────────────────────────────────────────────────

func makeReq(args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: args,
		},
	}
}

func TestRequiredFloat_returns_value_when_present(t *testing.T) {
	req := makeReq(map[string]any{"lat": 40.5})
	got, err := requiredFloat(req, "lat")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 40.5 {
		t.Errorf("got %v, want 40.5", got)
	}
}

func TestRequiredFloat_returns_error_when_missing(t *testing.T) {
	req := makeReq(map[string]any{})
	_, err := requiredFloat(req, "lat")
	if err == nil {
		t.Fatal("expected error for missing param")
	}
}

func TestRequiredFloat_returns_error_for_non_numeric(t *testing.T) {
	req := makeReq(map[string]any{"lat": true})
	_, err := requiredFloat(req, "lat")
	if err == nil {
		t.Fatal("expected error for non-numeric param")
	}
}

func TestRequiredFloat_coerces_string_numbers(t *testing.T) {
	req := makeReq(map[string]any{"lat": "40.5"})
	got, err := requiredFloat(req, "lat")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 40.5 {
		t.Errorf("got %v, want 40.5", got)
	}
}

func TestRequiredFloat_coerces_int(t *testing.T) {
	req := makeReq(map[string]any{"lat": int(10)})
	got, err := requiredFloat(req, "lat")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 10 {
		t.Errorf("got %v, want 10", got)
	}
}

// ── optionalFloat ────────────────────────────────────────────────────────────

func TestOptionalFloat_returns_value_when_present(t *testing.T) {
	req := makeReq(map[string]any{"radius_km": 25.0})
	got := optionalFloat(req, "radius_km", 10)
	if got != 25.0 {
		t.Errorf("got %v, want 25.0", got)
	}
}

func TestOptionalFloat_returns_default_when_absent(t *testing.T) {
	req := makeReq(map[string]any{})
	got := optionalFloat(req, "radius_km", 10)
	if got != 10 {
		t.Errorf("got %v, want 10", got)
	}
}

func TestOptionalFloat_returns_default_for_non_numeric(t *testing.T) {
	req := makeReq(map[string]any{"radius_km": true})
	got := optionalFloat(req, "radius_km", 10)
	if got != 10 {
		t.Errorf("got %v, want 10", got)
	}
}

func TestOptionalFloat_coerces_int(t *testing.T) {
	req := makeReq(map[string]any{"limit": int(20)})
	got := optionalFloat(req, "limit", 5)
	if got != 20 {
		t.Errorf("got %v, want 20", got)
	}
}

func TestOptionalFloat_coerces_string(t *testing.T) {
	req := makeReq(map[string]any{"limit": "15"})
	got := optionalFloat(req, "limit", 5)
	if got != 15 {
		t.Errorf("got %v, want 15", got)
	}
}
