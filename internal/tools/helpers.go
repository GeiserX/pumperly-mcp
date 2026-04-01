package tools

import (
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// requiredFloat extracts a required float64 argument from the request.
func requiredFloat(req mcp.CallToolRequest, key string) (float64, error) {
	v, ok := req.GetArguments()[key]
	if !ok {
		return 0, fmt.Errorf("missing required parameter: %s", key)
	}
	f, ok := v.(float64)
	if !ok {
		return 0, fmt.Errorf("parameter %s must be a number", key)
	}
	return f, nil
}

// optionalFloat extracts an optional float64 argument, returning a default if absent.
func optionalFloat(req mcp.CallToolRequest, key string, def float64) float64 {
	v, ok := req.GetArguments()[key]
	if !ok {
		return def
	}
	f, ok := v.(float64)
	if !ok {
		return def
	}
	return f
}
