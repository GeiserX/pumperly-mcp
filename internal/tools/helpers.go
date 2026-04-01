package tools

import (
	"encoding/json"
	"fmt"
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
	if !ok {
		return 0, fmt.Errorf("missing required parameter: %s", key)
	}
	f, ok := toFloat64(v)
	if !ok {
		return 0, fmt.Errorf("parameter %s must be a number", key)
	}
	return f, nil
}

// optionalFloat extracts an optional numeric argument, returning a default if absent.
func optionalFloat(req mcp.CallToolRequest, key string, def float64) float64 {
	v, ok := req.GetArguments()[key]
	if !ok {
		return def
	}
	f, ok := toFloat64(v)
	if !ok {
		return def
	}
	return f
}
