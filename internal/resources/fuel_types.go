package resources

import (
	"context"
	"encoding/json"

	"github.com/geiserx/pumperly-mcp/internal/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterFuelTypes wires the static pumperly://fuel-types list: every fuel
// code the tools accept, with a label, category and price unit.
func RegisterFuelTypes(s *server.MCPServer) {
	body, _ := json.Marshal(tools.FuelTypes)
	res := mcp.NewResource(
		"pumperly://fuel-types",
		"Fuel types",
		mcp.WithResourceDescription("The 16 fuel codes the station tools accept (B7 diesel, E5 petrol 95, EV charging...), with a label, category and the unit prices are quoted per (L, kg or kWh)."),
		mcp.WithMIMEType("application/json"),
	)
	s.AddResource(res, func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		return []mcp.ResourceContents{
			mcp.TextResourceContents{URI: "pumperly://fuel-types", MIMEType: "application/json", Text: string(body)},
		}, nil
	})
}
