package tools

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func makeReq(args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: args}}
}

func TestToFloat64(t *testing.T) {
	tests := []struct {
		input  any
		wantF  float64
		wantOK bool
	}{
		{float64(3.14), 3.14, true},
		{float32(2.5), 2.5, true},
		{int(42), 42, true},
		{int64(100), 100, true},
		{json.Number("9.81"), 9.81, true},
		{json.Number("notnum"), 0, false},
		{"123.456", 123.456, true},
		{"abc", 0, false},
		{true, 0, false},
		{nil, 0, false},
	}
	for _, tc := range tests {
		f, ok := toFloat64(tc.input)
		if ok != tc.wantOK || (ok && f != tc.wantF) {
			t.Errorf("toFloat64(%#v) = %v, %v", tc.input, f, ok)
		}
	}
}

func TestRequiredFloat(t *testing.T) {
	if v, err := requiredFloat(makeReq(map[string]any{"a": 1.5}), "a"); err != nil || v != 1.5 {
		t.Errorf("got %v, %v", v, err)
	}
	if _, err := requiredFloat(makeReq(map[string]any{}), "a"); err == nil {
		t.Error("missing: want error")
	}
	if _, err := requiredFloat(makeReq(map[string]any{"a": "NaN"}), "a"); err == nil {
		t.Error("NaN: want error")
	}
}

func TestOptionalFloat_never_falls_back_silently(t *testing.T) {
	if v, err := optionalFloat(makeReq(map[string]any{}), "r", 10, 0.5, 100); err != nil || v != 10 {
		t.Errorf("absent: %v, %v", v, err)
	}
	if v, err := optionalFloat(makeReq(map[string]any{"r": nil}), "r", 10, 0.5, 100); err != nil || v != 10 {
		t.Errorf("null: %v, %v", v, err)
	}
	if _, err := optionalFloat(makeReq(map[string]any{"r": "abc"}), "r", 10, 0.5, 100); err == nil {
		t.Error("non-numeric: want error, the old code returned the default")
	}
	if _, err := optionalFloat(makeReq(map[string]any{"r": 0.1}), "r", 10, 0.5, 100); err == nil {
		t.Error("below min: want error")
	}
	if v, err := optionalFloat(makeReq(map[string]any{"r": 100.0}), "r", 10, 0.5, 100); err != nil || v != 100 {
		t.Errorf("at max: %v, %v", v, err)
	}
}

func TestOptionalInt(t *testing.T) {
	if v, err := optionalInt(makeReq(map[string]any{"n": 7.0}), "n", 5, 1, 50); err != nil || v != 7 {
		t.Errorf("got %v, %v", v, err)
	}
	if _, err := optionalInt(makeReq(map[string]any{"n": 7.5}), "n", 5, 1, 50); err == nil {
		t.Error("fraction: want error")
	}
}

func TestHaversineKm(t *testing.T) {
	// Madrid to Valencia is about 302 km in a straight line.
	if d := haversineKm(40.4168, -3.7038, 39.4699, -0.3763); math.Abs(d-302) > 3 {
		t.Errorf("distance = %v", d)
	}
}

func TestResolvePlace_reads_lat_lon_text(t *testing.T) {
	for in, want := range map[string][2]float64{
		"40.4168,-3.7038":  {40.4168, -3.7038},
		" 40.4168 , -3.7 ": {40.4168, -3.7},
		"40.4168 -3.7038":  {40.4168, -3.7038},
		"-33.86;151.2":     {-33.86, 151.2},
	} {
		p, err := resolvePlace(nil, nil, in, "origin")
		if err != nil || p.Lat != want[0] || p.Lon != want[1] {
			t.Errorf("resolvePlace(%q) = %+v, %v", in, p, err)
		}
	}
}

func TestReadOnlyTool_sets_every_hint(t *testing.T) {
	tool := mcp.NewTool("x", readOnlyTool("X")...)
	a := tool.Annotations
	if tool.Title != "X" || a.Title != "X" || !*a.ReadOnlyHint || *a.DestructiveHint || !*a.IdempotentHint || !*a.OpenWorldHint {
		t.Errorf("annotations = %+v", a)
	}
}
