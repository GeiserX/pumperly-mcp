package tools

import (
	"fmt"
	"strings"
	"testing"
)

const routeFixture = `{"routes":[{"geometry":{"type":"LineString","coordinates":[[-3.7038,40.4168],[-2.0,40.0],[-0.3763,39.4699]]},"distance":358.916,"duration":11683.052,"bbox":[-3.7038,39.4699,-0.3763,40.4168],"durations":[0,5000,11683]}]}`

func TestCalculateRoute_posts_lon_lat_order_and_summarises(t *testing.T) {
	api := newFakeAPI(t, map[string]string{"POST /api/route": routeFixture})
	_, h := NewCalculateRoute(api.client())

	res := as[RouteResult](t, call(t, h, map[string]any{
		"origin_lat": 40.4168, "origin_lon": -3.7038,
		"dest_lat": 39.4699, "dest_lon": -0.3763,
		"waypoints": []any{[]any{-1.13, 40.34}},
	}))

	body := api.last(t, "/api/route").Body
	if got := fmt.Sprint(body["origin"], body["destination"], body["waypoints"]); got != "[-3.7038 40.4168] [-0.3763 39.4699] [[-1.13 40.34]]" {
		t.Errorf("body = %s", got)
	}
	r := res.Routes[0]
	if r.DistanceKm != 358.9 || r.DurationMin != 195 || r.Points != 3 || r.Geometry != nil {
		t.Errorf("route = %+v", r)
	}
}

func TestCalculateRoute_includes_geometry_on_request(t *testing.T) {
	api := newFakeAPI(t, map[string]string{"POST /api/route": routeFixture})
	_, h := NewCalculateRoute(api.client())
	res := as[RouteResult](t, call(t, h, map[string]any{
		"origin": "40.4168,-3.7038", "destination": "39.4699, -0.3763", "include_geometry": true,
	}))
	if g := res.Routes[0].Geometry; g == nil || g.Type != "LineString" || len(g.Coordinates) != 3 {
		t.Errorf("geometry = %+v", g)
	}
	if body := api.last(t, "/api/route").Body; fmt.Sprint(body["origin"]) != "[-3.7038 40.4168]" {
		t.Errorf("\"lat,lon\" text was not turned into [lon, lat]: %v", body["origin"])
	}
	if _, ok := api.last(t, "/api/route").Body["waypoints"]; ok {
		t.Error("empty waypoints sent")
	}
}

func TestCalculateRoute_geocodes_place_names(t *testing.T) {
	api := newFakeAPI(t, map[string]string{"GET /api/geocode": geocodeFixture, "POST /api/route": routeFixture})
	_, h := NewCalculateRoute(api.client())
	res := as[RouteResult](t, call(t, h, map[string]any{"origin": "Atocha", "dest_lat": 39.47, "dest_lon": -0.38}))
	if res.Origin.Name != "Atocha" || fmt.Sprint(api.last(t, "/api/route").Body["origin"]) != "[-3.688582 40.406975]" {
		t.Errorf("origin = %+v", res.Origin)
	}
}

func TestCalculateRoute_rejects_bad_input(t *testing.T) {
	api := newFakeAPI(t, map[string]string{"POST /api/route": routeFixture})
	_, h := NewCalculateRoute(api.client())
	six := []any{}
	for i := 0; i < 6; i++ {
		six = append(six, []any{-3.0, 40.0})
	}
	for name, tc := range map[string]struct {
		args map[string]any
		want string
	}{
		"no origin":         {map[string]any{"dest_lat": 39.0, "dest_lon": -0.3}, "give origin"},
		"half origin":       {map[string]any{"origin_lat": 40.0, "dest_lat": 39.0, "dest_lon": -0.3}, "both origin_lat and origin_lon"},
		"six waypoints":     {map[string]any{"origin": "40,-3", "destination": "39,-0.3", "waypoints": six}, "at most 5 waypoints"},
		"lat,lon waypoint":  {map[string]any{"origin": "40,-3", "destination": "39,-0.3", "waypoints": []any{[]any{40.0, 120.0}}}, "longitude first"},
		"waypoint not pair": {map[string]any{"origin": "40,-3", "destination": "39,-0.3", "waypoints": []any{[]any{1.0}}}, "[lon, lat] pair"},
		"bad lat,lon text":  {map[string]any{"origin": "95,-3", "destination": "39,-0.3"}, "latitude must be between"},
	} {
		t.Run(name, func(t *testing.T) { wantError(t, call(t, h, tc.args), tc.want) })
	}
	if n := len(api.requests()); n != 0 {
		t.Errorf("%d upstream requests for invalid input", n)
	}
}

func TestCalculateRoute_api_error_is_reported(t *testing.T) {
	api := newFakeAPI(t, map[string]string{})
	_, h := NewCalculateRoute(api.client())
	res := call(t, h, map[string]any{"origin": "40,-3", "destination": "39,-0.3"})
	if !res.IsError || !strings.Contains(text(t, res), "HTTP 404") {
		t.Errorf("result = %s", text(t, res))
	}
}
