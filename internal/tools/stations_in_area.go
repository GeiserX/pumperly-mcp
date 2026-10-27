package tools

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/geiserx/pumperly-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// NewGetStationsInArea builds the Tool definition plus its handler.
func NewGetStationsInArea(c *client.Client) (mcp.Tool, server.ToolHandlerFunc) {
	opts := append(readOnlyTool("Fuel or EV stations in an area"),
		mcp.WithDescription("List the fuel stations or EV chargers inside a bounding box with current prices, cheapest first by default, plus the min, median and max price per currency. "+
			"Use it for \"cheapest diesel in this town\". For a point and radius use find_nearest_stations; along a trip use cheapest_fuel_on_route. "+
			"Prices are per litre (per kg for CNG, LNG and H2, per kWh for EV) in local currency. Keep the box city-sized: a whole-country box can exceed the 8 MB answer cap."),
		mcp.WithString("bbox",
			mcp.Required(),
			mcp.Pattern(`^\s*-?\d+(\.\d+)?\s*,\s*-?\d+(\.\d+)?\s*,\s*-?\d+(\.\d+)?\s*,\s*-?\d+(\.\d+)?\s*$`),
			mcp.Description("Bounding box as \"minLon,minLat,maxLon,maxLat\" in decimal degrees, longitude first. Example for central Madrid: \"-3.75,40.38,-3.65,40.45\"."),
		),
		fuelParam(),
		mcp.WithInteger("limit", mcp.Min(1), mcp.Max(100), mcp.DefaultNumber(20),
			mcp.Description("How many stations to return, 1 to 100. Default 20.")),
		mcp.WithString("sort_by", mcp.Enum("price", "distance"), mcp.DefaultString("price"),
			mcp.Description("price (cheapest first, default) or distance (from the centre of the box).")),
		mcp.WithOutputSchema[StationList](),
	)
	tool := mcp.NewTool("get_stations_in_area", opts...)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		bboxText, err := req.RequireString("bbox")
		if err != nil {
			return toolError(err)
		}
		bbox, err := parseBBox(bboxText)
		if err != nil {
			return toolError(err)
		}
		fuel, err := requireFuel(req)
		if err != nil {
			return toolError(err)
		}
		limit, err := optionalInt(req, "limit", 20, 1, 100)
		if err != nil {
			return toolError(err)
		}
		sortBy, err := optionalEnum(req, "sort_by", "price", "price", "distance")
		if err != nil {
			return toolError(err)
		}

		body, err := c.GetStationsInArea(ctx, formatBBox(bbox), fuel)
		if err != nil {
			return toolError(err)
		}
		st, err := decodeStations(body, fuel, 0)
		if err != nil {
			return toolError(err)
		}
		cLat, cLon := (bbox[1]+bbox[3])/2, (bbox[0]+bbox[2])/2
		for i := range st {
			d := round(haversineKm(cLat, cLon, st[i].Lat, st[i].Lon), 2)
			st[i].DistanceKm = &d
		}
		return structured(newStationList(st, fuel, sortBy, limit))
	}

	return tool, handler
}

// parseBBox reads "minLon,minLat,maxLon,maxLat" and checks its ranges and
// order, so a swapped box fails here with a clear message.
func parseBBox(s string) ([4]float64, error) {
	var b [4]float64
	parts := strings.Split(s, ",")
	if len(parts) != 4 {
		return b, fmt.Errorf("bbox must be four numbers \"minLon,minLat,maxLon,maxLat\", got %q", s)
	}
	for i, p := range parts {
		f, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil {
			return b, fmt.Errorf("bbox value %q is not a number", strings.TrimSpace(p))
		}
		b[i] = f
	}
	if err := checkLatLon(b[1], b[0]); err != nil {
		return b, fmt.Errorf("bbox min corner: %v", err)
	}
	if err := checkLatLon(b[3], b[2]); err != nil {
		return b, fmt.Errorf("bbox max corner: %v", err)
	}
	if b[0] >= b[2] || b[1] >= b[3] {
		return b, fmt.Errorf("bbox must be minLon,minLat,maxLon,maxLat with min < max (longitude first)")
	}
	return b, nil
}

func formatBBox(b [4]float64) string {
	parts := make([]string, 4)
	for i, f := range b {
		parts[i] = strconv.FormatFloat(f, 'f', -1, 64)
	}
	return strings.Join(parts, ",")
}
