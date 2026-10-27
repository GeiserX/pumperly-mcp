package tools

import (
	"context"
	"fmt"

	"github.com/geiserx/pumperly-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// upstreamMaxNearest is the most stations /api/stations/nearest returns.
const upstreamMaxNearest = 50

// NewFindNearestStations builds the Tool definition plus its handler.
func NewFindNearestStations(c *client.Client) (mcp.Tool, server.ToolHandlerFunc) {
	opts := append(readOnlyTool("Find nearest fuel or EV stations"),
		mcp.WithDescription("Find the fuel stations or EV chargers closest to a point, with current prices. "+
			"Give the point as lat and lon, or as a place name in near (geocoded for you). "+
			"sort_by=distance (default) returns the nearest; sort_by=price returns the cheapest among the 50 nearest within the radius. "+
			"Prices are per litre (per kg for CNG, LNG and H2, per kWh for EV) in the station's local currency. "+
			"Covers about 40 countries, mostly in Europe; pumperly://stats lists them."),
		mcp.WithNumber("lat", mcp.Min(-90), mcp.Max(90),
			mcp.Description("Latitude of the search point in decimal degrees, e.g. 40.4168. Give with lon, or use near instead.")),
		mcp.WithNumber("lon", mcp.Min(-180), mcp.Max(180),
			mcp.Description("Longitude of the search point in decimal degrees, e.g. -3.7038. Give with lat, or use near instead.")),
		mcp.WithString("near", mcp.MinLength(1), mcp.MaxLength(200),
			mcp.Description("Place name or address to search around, e.g. \"Atocha, Madrid\". Used when lat and lon are not given.")),
		fuelParam(),
		mcp.WithNumber("radius_km", mcp.Min(0.5), mcp.Max(100), mcp.DefaultNumber(10.0),
			mcp.Description("Search radius in km, 0.5 to 100. Default 10.")),
		mcp.WithInteger("limit", mcp.Min(1), mcp.Max(upstreamMaxNearest), mcp.DefaultNumber(5),
			mcp.Description("How many stations to return, 1 to 50. Default 5.")),
		mcp.WithString("sort_by", mcp.Enum("distance", "price"), mcp.DefaultString("distance"),
			mcp.Description("distance (nearest first, default) or price (cheapest first).")),
		mcp.WithOutputSchema[StationList](),
	)
	tool := mcp.NewTool("find_nearest_stations", opts...)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		center, err := searchPoint(ctx, c, req)
		if err != nil {
			return toolError(err)
		}
		fuel, err := requireFuel(req)
		if err != nil {
			return toolError(err)
		}
		radiusKm, err := optionalFloat(req, "radius_km", 10, 0.5, 100)
		if err != nil {
			return toolError(err)
		}
		limit, err := optionalInt(req, "limit", 5, 1, upstreamMaxNearest)
		if err != nil {
			return toolError(err)
		}
		sortBy, err := optionalEnum(req, "sort_by", "distance", "distance", "price")
		if err != nil {
			return toolError(err)
		}

		// The API returns the nearest N. To find the cheapest, ask for all
		// it will give and sort here.
		fetch := limit
		if sortBy == "price" {
			fetch = upstreamMaxNearest
		}
		body, err := c.FindNearestStations(ctx, center.Lat, center.Lon, fuel, radiusKm, fetch)
		if err != nil {
			return toolError(err)
		}
		st, err := decodeStations(body, fuel, 0)
		if err != nil {
			return toolError(err)
		}
		list := newStationList(st, fuel, sortBy, limit)
		list.Center = &center
		return structured(list)
	}

	return tool, handler
}

// searchPoint reads lat/lon, or geocodes near.
func searchPoint(ctx context.Context, c *client.Client, req mcp.CallToolRequest) (Place, error) {
	lat, lon, ok, err := latLon(req, "lat", "lon")
	if err != nil {
		return Place{}, err
	}
	if ok {
		return Place{Name: fmt.Sprintf("%g,%g", lat, lon), Lat: lat, Lon: lon}, nil
	}
	near, err := optionalString(req, "near")
	if err != nil {
		return Place{}, err
	}
	if near == "" {
		return Place{}, fmt.Errorf("give lat and lon, or a place name in near")
	}
	return resolvePlace(ctx, c, near, "near")
}
