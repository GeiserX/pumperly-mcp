package tools

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"
)

// toFloat64 coerces a value to float64, accepting float64, int-family, and
// numeric strings — matching the range of representations MCP clients may send.
func toFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	case string:
		f, err := strconv.ParseFloat(n, 64)
		return f, err == nil
	default:
		return 0, false
	}
}

// requiredFloat extracts a required numeric argument from the request.
func requiredFloat(req mcp.CallToolRequest, key string) (float64, error) {
	v, ok := req.GetArguments()[key]
	if !ok || v == nil {
		return 0, fmt.Errorf("missing required parameter: %s", key)
	}
	f, ok := toFloat64(v)
	if !ok || math.IsNaN(f) || math.IsInf(f, 0) {
		return 0, fmt.Errorf("parameter %s must be a number", key)
	}
	return f, nil
}

// optionalFloat extracts an optional numeric argument within [min, max],
// returning def when it is absent. A present but invalid value is an error,
// never a silent fallback to the default.
func optionalFloat(req mcp.CallToolRequest, key string, def, min, max float64) (float64, error) {
	if v, ok := req.GetArguments()[key]; !ok || v == nil {
		return def, nil
	}
	f, err := requiredFloat(req, key)
	if err != nil {
		return 0, err
	}
	if f < min || f > max {
		return 0, fmt.Errorf("parameter %s must be between %g and %g, got %g", key, min, max, f)
	}
	return f, nil
}

// optionalInt is optionalFloat for whole numbers.
func optionalInt(req mcp.CallToolRequest, key string, def, min, max int) (int, error) {
	f, err := optionalFloat(req, key, float64(def), float64(min), float64(max))
	if err != nil {
		return 0, err
	}
	if f != math.Trunc(f) {
		return 0, fmt.Errorf("parameter %s must be a whole number, got %g", key, f)
	}
	return int(f), nil
}

// optionalString returns a string argument or "" when absent.
func optionalString(req mcp.CallToolRequest, key string) (string, error) {
	v, ok := req.GetArguments()[key]
	if !ok || v == nil {
		return "", nil
	}
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("parameter %s must be a string", key)
	}
	return s, nil
}

// optionalEnum returns a string argument that must be one of allowed, or def.
func optionalEnum(req mcp.CallToolRequest, key, def string, allowed ...string) (string, error) {
	s, err := optionalString(req, key)
	if err != nil || s == "" {
		return def, err
	}
	for _, a := range allowed {
		if s == a {
			return s, nil
		}
	}
	return "", fmt.Errorf("parameter %s must be one of %v, got %q", key, allowed, s)
}

// latLon reads a lat/lon pair. Both or neither may be given; ok reports
// whether they were.
func latLon(req mcp.CallToolRequest, latKey, lonKey string) (lat, lon float64, ok bool, err error) {
	args := req.GetArguments()
	_, hasLat := args[latKey]
	_, hasLon := args[lonKey]
	if hasLat != hasLon {
		return 0, 0, false, fmt.Errorf("give both %s and %s, or neither", latKey, lonKey)
	}
	if !hasLat {
		return 0, 0, false, nil
	}
	if lat, err = requiredFloat(req, latKey); err != nil {
		return 0, 0, false, err
	}
	if lon, err = requiredFloat(req, lonKey); err != nil {
		return 0, 0, false, err
	}
	if err := checkLatLon(lat, lon); err != nil {
		return 0, 0, false, err
	}
	return lat, lon, true, nil
}

func checkLatLon(lat, lon float64) error {
	if lat < -90 || lat > 90 {
		return fmt.Errorf("latitude must be between -90 and 90, got %g", lat)
	}
	if lon < -180 || lon > 180 {
		return fmt.Errorf("longitude must be between -180 and 180, got %g", lon)
	}
	return nil
}

// haversineKm is the great-circle distance between two points.
func haversineKm(lat1, lon1, lat2, lon2 float64) float64 {
	const r = 6371.0088
	rad := math.Pi / 180
	dLat := (lat2 - lat1) * rad
	dLon := (lon2 - lon1) * rad
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1*rad)*math.Cos(lat2*rad)*math.Sin(dLon/2)*math.Sin(dLon/2)
	return 2 * r * math.Asin(math.Min(1, math.Sqrt(a)))
}

func round(v float64, places int) float64 {
	p := math.Pow(10, float64(places))
	return math.Round(v*p) / p
}

// readOnlyTool marks a tool as a pure query of the public Pumperly API: it
// reads, never writes, and talks to a service outside this process.
func readOnlyTool(title string) []mcp.ToolOption {
	return []mcp.ToolOption{
		mcp.WithToolTitle(title),
		mcp.WithTitleAnnotation(title),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(true),
	}
}

// structured returns v as structuredContent plus the same JSON as text, for
// clients that only read text content.
func structured(v any) (*mcp.CallToolResult, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return mcp.NewToolResultError("could not encode the result"), nil
	}
	return mcp.NewToolResultStructured(v, string(b)), nil
}

func toolError(err error) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultError(err.Error()), nil
}
