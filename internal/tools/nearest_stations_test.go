package tools

import (
	"testing"
)

func TestFindNearestStations_sends_exact_query_and_defaults(t *testing.T) {
	api := newFakeAPI(t, map[string]string{"GET /api/stations/nearest": nearestFixture})
	_, h := NewFindNearestStations(api.client())

	res := as[StationList](t, call(t, h, map[string]any{"lat": 40.4168, "lon": -3.7038, "fuel": "B7"}))

	q := api.last(t, "/api/stations/nearest").Query
	for k, want := range map[string]string{"lat": "40.4168", "lon": "-3.7038", "fuel": "B7", "radius_km": "10", "limit": "5"} {
		if q[k] != want {
			t.Errorf("%s = %q, want %q", k, q[k], want)
		}
	}
	if res.SortBy != "distance" || res.Unit != "L" || res.Found != 4 || res.Returned != 4 {
		t.Errorf("list = %+v", res)
	}
	if res.Stations[0].Name != "No Price" || *res.Stations[0].DistanceKm != 0.5 {
		t.Errorf("not sorted by distance: %+v", res.Stations[0])
	}
	if res.Center == nil || res.Center.Lat != 40.4168 {
		t.Errorf("center = %+v", res.Center)
	}
}

func TestFindNearestStations_sort_by_price_asks_for_50_and_returns_the_cheapest(t *testing.T) {
	api := newFakeAPI(t, map[string]string{"GET /api/stations/nearest": nearestFixture})
	_, h := NewFindNearestStations(api.client())

	res := as[StationList](t, call(t, h, map[string]any{"lat": 40.4, "lon": -3.7, "fuel": "B7", "sort_by": "price", "limit": 2}))

	if got := api.last(t, "/api/stations/nearest").Query["limit"]; got != "50" {
		t.Errorf("upstream limit = %q, want 50", got)
	}
	if res.Returned != 2 || res.Found != 4 {
		t.Fatalf("returned %d of %d", res.Returned, res.Found)
	}
	if res.Stations[0].Name != "Ballenoil" || res.Stations[1].Name != "Blanca Madrid" {
		t.Errorf("order = %s, %s", res.Stations[0].Name, res.Stations[1].Name)
	}
	if len(res.Prices) != 1 || res.Prices[0].Min != 1.659 || res.Prices[0].Max != 1.995 || res.Prices[0].Count != 3 {
		t.Errorf("prices = %+v", res.Prices)
	}
}

func TestFindNearestStations_geocodes_near(t *testing.T) {
	api := newFakeAPI(t, map[string]string{
		"GET /api/geocode":          geocodeFixture,
		"GET /api/stations/nearest": nearestFixture,
	})
	_, h := NewFindNearestStations(api.client())

	res := as[StationList](t, call(t, h, map[string]any{"near": "Atocha, Madrid", "fuel": "B7"}))

	if got := api.last(t, "/api/geocode").Query["q"]; got != "Atocha, Madrid" {
		t.Errorf("geocode q = %q", got)
	}
	q := api.last(t, "/api/stations/nearest").Query
	if q["lat"] != "40.406975" || q["lon"] != "-3.688582" {
		t.Errorf("searched at %s,%s, want the geocoded point", q["lat"], q["lon"])
	}
	if res.Center.Name != "Atocha" {
		t.Errorf("center = %+v", res.Center)
	}
}

func TestFindNearestStations_rejects_bad_input_before_calling_the_api(t *testing.T) {
	api := newFakeAPI(t, map[string]string{"GET /api/stations/nearest": nearestFixture})
	_, h := NewFindNearestStations(api.client())

	for name, tc := range map[string]struct {
		args map[string]any
		want string
	}{
		"no point":            {map[string]any{"fuel": "B7"}, "give lat and lon"},
		"lat without lon":     {map[string]any{"lat": 40.0, "fuel": "B7"}, "both lat and lon"},
		"lat out of range":    {map[string]any{"lat": 140.0, "lon": 1.0, "fuel": "B7"}, "latitude must be between"},
		"E85 is not a code":   {map[string]any{"lat": 40.0, "lon": -3.0, "fuel": "E85"}, "unknown fuel"},
		"radius too big":      {map[string]any{"lat": 40.0, "lon": -3.0, "fuel": "B7", "radius_km": 500.0}, "radius_km must be between"},
		"radius not a number": {map[string]any{"lat": 40.0, "lon": -3.0, "fuel": "B7", "radius_km": "far"}, "radius_km must be a number"},
		"limit fraction":      {map[string]any{"lat": 40.0, "lon": -3.0, "fuel": "B7", "limit": 2.5}, "whole number"},
		"bad sort":            {map[string]any{"lat": 40.0, "lon": -3.0, "fuel": "B7", "sort_by": "name"}, "sort_by must be one of"},
	} {
		t.Run(name, func(t *testing.T) {
			wantError(t, call(t, h, tc.args), tc.want)
		})
	}
	if n := len(api.requests()); n != 0 {
		t.Errorf("%d upstream requests for invalid input", n)
	}
}

func TestFindNearestStations_api_error_is_a_short_tool_error(t *testing.T) {
	api := newFakeAPI(t, map[string]string{"GET /api/stations/nearest": "!"})
	_, h := NewFindNearestStations(api.client())
	wantError(t, call(t, h, map[string]any{"lat": 40.0, "lon": -3.0, "fuel": "B7"}), "HTTP 400): Invalid parameters (fuel: Invalid option)")
}

func TestFindNearestStations_rejects_a_non_geojson_answer(t *testing.T) {
	api := newFakeAPI(t, map[string]string{"GET /api/stations/nearest": `[{"name":"x"}]`})
	_, h := NewFindNearestStations(api.client())
	wantError(t, call(t, h, map[string]any{"lat": 40.0, "lon": -3.0, "fuel": "B7"}), "not a station list")
}
