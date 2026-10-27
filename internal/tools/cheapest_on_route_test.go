package tools

import (
	"fmt"
	"strings"
	"testing"
)

func TestCheapestFuelOnRoute_routes_then_searches_in_one_call(t *testing.T) {
	// A 2500-point route: the corridor search must get at most 2000.
	var coords []string
	for i := 0; i < 2500; i++ {
		coords = append(coords, fmt.Sprintf("[%g,%g]", -3.7+float64(i)*0.0013, 40.4-float64(i)*0.0004))
	}
	longRoute := `{"routes":[{"geometry":{"type":"LineString","coordinates":[` + strings.Join(coords, ",") + `]},"distance":350,"duration":12000,"bbox":[-3.7,39.4,-0.4,40.4]}]}`
	api := newFakeAPI(t, map[string]string{
		"GET /api/geocode":         geocodeFixture,
		"POST /api/route":          longRoute,
		"POST /api/route-stations": routeStationsFixture,
	})
	_, h := NewCheapestFuelOnRoute(api.client())

	res := as[RouteFuelResult](t, call(t, h, map[string]any{"origin": "Atocha", "destination": "39.4699,-0.3763", "fuel": "b7", "limit": 2}))

	rs := api.last(t, "/api/route-stations").Body
	geom, _ := rs["geometry"].(map[string]any)
	pts, _ := geom["coordinates"].([]any)
	if len(pts) != 2000 {
		t.Errorf("corridor search got %d points, want 2000", len(pts))
	}
	if rs["corridorKm"] != 2.0 || rs["fuel"] != "B7" {
		t.Errorf("route-stations body = corridorKm %v fuel %v", rs["corridorKm"], rs["fuel"])
	}
	if fmt.Sprint(api.last(t, "/api/route").Body["destination"]) != "[-0.3763 39.4699]" {
		t.Errorf("destination = %v", api.last(t, "/api/route").Body["destination"])
	}
	if res.DistanceKm != 350 || res.DurationMin != 200 || res.Origin.Name != "Atocha" {
		t.Errorf("result = %+v", res)
	}
	st := res.Stations
	if st.Returned != 2 || st.Found != 3 || st.Stations[0].Name != "Start" || st.Stations[1].Name != "Mid" {
		t.Errorf("stations = %+v", st)
	}
	if km := st.Stations[1].RouteKm; km == nil || *km != 175 {
		t.Errorf("route_km = %v, want 175 (0.5 of 350 km)", km)
	}
}

func TestCheapestFuelOnRoute_reports_an_unknown_place(t *testing.T) {
	api := newFakeAPI(t, map[string]string{"GET /api/geocode": `[]`})
	_, h := NewCheapestFuelOnRoute(api.client())
	wantError(t, call(t, h, map[string]any{"origin": "Nowhereville", "destination": "Madrid", "fuel": "B7"}), `no place found for origin "Nowhereville"`)
}

func TestCheapestFuelOnRoute_needs_a_fuel(t *testing.T) {
	api := newFakeAPI(t, nil)
	_, h := NewCheapestFuelOnRoute(api.client())
	wantError(t, call(t, h, map[string]any{"origin": "40,-3", "destination": "39,-0.3"}), "fuel")
}
