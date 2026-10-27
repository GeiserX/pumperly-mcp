package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/geiserx/pumperly-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// maxWaypoints is the API's limit (route.ts bodySchema waypoints max(5)).
const maxWaypoints = 5

// LineString is a GeoJSON line of [lon, lat] points.
type LineString struct {
	Type        string      `json:"type" jsonschema:"enum=LineString"`
	Coordinates [][]float64 `json:"coordinates"`
}

// RouteSummary is one route option.
type RouteSummary struct {
	DistanceKm  float64 `json:"distance_km"`
	DurationMin float64 `json:"duration_min"`
	// Bbox is [minLon, minLat, maxLon, maxLat].
	Bbox     []float64   `json:"bbox,omitempty"`
	Points   int         `json:"points" jsonschema:"Number of points in the route line"`
	Geometry *LineString `json:"geometry,omitempty"`
}

// RouteResult is the structured result of calculate_route.
type RouteResult struct {
	Origin      Place          `json:"origin"`
	Destination Place          `json:"destination"`
	Routes      []RouteSummary `json:"routes"`
}

type apiRoute struct {
	Geometry LineString `json:"geometry"`
	Distance float64    `json:"distance"` // km
	Duration float64    `json:"duration"` // s
	Bbox     []float64  `json:"bbox"`
}

func decodeRoutes(body []byte) ([]apiRoute, error) {
	var r struct {
		Routes []apiRoute `json:"routes"`
	}
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("unexpected answer from the Pumperly API (not a route)")
	}
	if len(r.Routes) == 0 {
		return nil, fmt.Errorf("no route found between these points")
	}
	return r.Routes, nil
}

func routeEndParams() []mcp.ToolOption {
	return []mcp.ToolOption{
		mcp.WithString("origin", mcp.MinLength(1), mcp.MaxLength(200),
			mcp.Description("Start as a place name (\"Madrid\") or \"lat,lon\" text (\"40.4168,-3.7038\"). Or give origin_lat and origin_lon.")),
		mcp.WithNumber("origin_lat", mcp.Min(-90), mcp.Max(90), mcp.Description("Start latitude, decimal degrees.")),
		mcp.WithNumber("origin_lon", mcp.Min(-180), mcp.Max(180), mcp.Description("Start longitude, decimal degrees.")),
		mcp.WithString("destination", mcp.MinLength(1), mcp.MaxLength(200),
			mcp.Description("End as a place name (\"Valencia\") or \"lat,lon\" text. Or give dest_lat and dest_lon.")),
		mcp.WithNumber("dest_lat", mcp.Min(-90), mcp.Max(90), mcp.Description("End latitude, decimal degrees.")),
		mcp.WithNumber("dest_lon", mcp.Min(-180), mcp.Max(180), mcp.Description("End longitude, decimal degrees.")),
	}
}

// routeEnd reads one end of a route: a lat/lon pair, else the text param.
func routeEnd(ctx context.Context, c *client.Client, req mcp.CallToolRequest, textKey, latKey, lonKey string) (Place, error) {
	lat, lon, ok, err := latLon(req, latKey, lonKey)
	if err != nil {
		return Place{}, err
	}
	if ok {
		return Place{Name: fmt.Sprintf("%g,%g", lat, lon), Lat: lat, Lon: lon}, nil
	}
	text, err := optionalString(req, textKey)
	if err != nil {
		return Place{}, err
	}
	if text == "" {
		return Place{}, fmt.Errorf("give %s (a place name or \"lat,lon\"), or %s and %s", textKey, latKey, lonKey)
	}
	return resolvePlace(ctx, c, text, textKey)
}

// NewCalculateRoute builds the Tool definition plus its handler.
func NewCalculateRoute(c *client.Client) (mcp.Tool, server.ToolHandlerFunc) {
	opts := append(readOnlyTool("Calculate a driving route"), routeEndParams()...)
	opts = append(opts,
		mcp.WithDescription("Calculate a driving route and return its distance (km), driving time (minutes) and bounding box, with up to 2 alternatives when there are no waypoints. "+
			"Ends can be place names, \"lat,lon\" text or coordinates. "+
			"To find fuel along the way, call cheapest_fuel_on_route instead: it routes and searches in one call. "+
			"Set include_geometry only to pass the line to find_route_stations; it can be up to 2000 points."),
		mcp.WithArray("waypoints",
			mcp.MaxItems(maxWaypoints),
			mcp.Items(map[string]any{
				"type":     "array",
				"items":    map[string]any{"type": "number"},
				"minItems": 2,
				"maxItems": 2,
			}),
			mcp.Description("Optional stops in order, as [lon, lat] pairs (longitude first), at most 5. Example: [[-1.13, 40.34]]."),
		),
		mcp.WithBoolean("include_geometry", mcp.DefaultBool(false),
			mcp.Description("Include each route's GeoJSON LineString ([lon, lat] points). Default false.")),
		mcp.WithOutputSchema[RouteResult](),
	)
	tool := mcp.NewTool("calculate_route", opts...)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		origin, err := routeEnd(ctx, c, req, "origin", "origin_lat", "origin_lon")
		if err != nil {
			return toolError(err)
		}
		dest, err := routeEnd(ctx, c, req, "destination", "dest_lat", "dest_lon")
		if err != nil {
			return toolError(err)
		}
		waypoints, err := parseWaypoints(req.GetArguments()["waypoints"])
		if err != nil {
			return toolError(err)
		}
		body := routeRequest{
			Origin:      [2]float64{origin.Lon, origin.Lat},
			Destination: [2]float64{dest.Lon, dest.Lat},
			Waypoints:   waypoints,
		}
		resp, err := c.CalculateRoute(ctx, body)
		if err != nil {
			return toolError(err)
		}
		routes, err := decodeRoutes(resp)
		if err != nil {
			return toolError(err)
		}
		withGeometry := req.GetBool("include_geometry", false)
		out := RouteResult{Origin: origin, Destination: dest}
		for _, r := range routes {
			s := RouteSummary{
				DistanceKm:  round(r.Distance, 1),
				DurationMin: round(r.Duration/60, 0),
				Bbox:        r.Bbox,
				Points:      len(r.Geometry.Coordinates),
			}
			if withGeometry {
				g := r.Geometry
				s.Geometry = &g
			}
			out.Routes = append(out.Routes, s)
		}
		return structured(out)
	}

	return tool, handler
}

type routeRequest struct {
	Origin      [2]float64   `json:"origin"`
	Destination [2]float64   `json:"destination"`
	Waypoints   [][2]float64 `json:"waypoints,omitempty"`
}

// parseWaypoints reads an array of [lon, lat] pairs.
func parseWaypoints(v any) ([][2]float64, error) {
	if v == nil {
		return nil, nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("waypoints must be an array of [lon, lat] pairs")
	}
	if len(arr) > maxWaypoints {
		return nil, fmt.Errorf("at most %d waypoints, got %d", maxWaypoints, len(arr))
	}
	out := make([][2]float64, 0, len(arr))
	for i, item := range arr {
		pair, ok := item.([]any)
		if !ok || len(pair) != 2 {
			return nil, fmt.Errorf("waypoint %d must be a [lon, lat] pair", i+1)
		}
		lon, ok1 := toFloat64(pair[0])
		lat, ok2 := toFloat64(pair[1])
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("waypoint %d must be two numbers [lon, lat]", i+1)
		}
		if err := checkLatLon(lat, lon); err != nil {
			return nil, fmt.Errorf("waypoint %d: %v (pairs are [lon, lat], longitude first)", i+1, err)
		}
		out = append(out, [2]float64{lon, lat})
	}
	return out, nil
}
