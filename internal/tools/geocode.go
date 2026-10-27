package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/geiserx/pumperly-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// GeocodeResult is the structured result of geocode.
type GeocodeResult struct {
	Query   string  `json:"query"`
	Results []Place `json:"results"`
}

// NewGeocode builds the Tool definition plus its handler.
func NewGeocode(c *client.Client) (mcp.Tool, server.ToolHandlerFunc) {
	opts := append(readOnlyTool("Geocode a place name"),
		mcp.WithDescription("Turn a place name or address into coordinates (lat, lon), best match first. "+
			"The station and route tools already accept place names, so call this only to pick between ambiguous places or to get coordinates for a bbox. "+
			"Give lat and lon to prefer results near a point."),
		mcp.WithString("query",
			mcp.Required(),
			mcp.MinLength(1),
			mcp.MaxLength(200),
			mcp.Description("City, address or point of interest, e.g. \"Puerta del Sol, Madrid\". At most 200 characters."),
		),
		mcp.WithNumber("lat", mcp.Min(-90), mcp.Max(90),
			mcp.Description("Optional latitude to prefer nearby results. Give with lon.")),
		mcp.WithNumber("lon", mcp.Min(-180), mcp.Max(180),
			mcp.Description("Optional longitude to prefer nearby results. Give with lat.")),
		mcp.WithInteger("limit", mcp.Min(1), mcp.Max(10), mcp.DefaultNumber(5),
			mcp.Description("How many matches to return, 1 to 10. Default 5.")),
		mcp.WithOutputSchema[GeocodeResult](),
	)
	tool := mcp.NewTool("geocode", opts...)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := req.RequireString("query")
		if err != nil {
			return toolError(err)
		}
		query = strings.TrimSpace(query)
		if query == "" || len([]rune(query)) > 200 {
			return toolError(fmt.Errorf("query must be 1 to 200 characters"))
		}
		lat, lon, bias, err := latLon(req, "lat", "lon")
		if err != nil {
			return toolError(err)
		}
		limit, err := optionalInt(req, "limit", 5, 1, 10)
		if err != nil {
			return toolError(err)
		}
		var latP, lonP *float64
		if bias {
			latP, lonP = &lat, &lon
		}
		body, err := c.Geocode(ctx, query, latP, lonP)
		if err != nil {
			return toolError(err)
		}
		places, err := decodePlaces(body)
		if err != nil {
			return toolError(err)
		}
		if len(places) > limit {
			places = places[:limit]
		}
		return structured(GeocodeResult{Query: query, Results: places})
	}

	return tool, handler
}
