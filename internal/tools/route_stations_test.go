package tools

import (
	"fmt"
	"testing"
)

const routeStationsFixture = `{"type":"FeatureCollection","features":[
{"type":"Feature","geometry":{"type":"Point","coordinates":[-2.0,40.0]},"properties":{"name":"Mid","brand":"Repsol","address":"A-3 km 100","city":"Tarancón","country":"ES","fuelType":"B7","currency":"EUR","price":1.70,"routeFraction":0.5}},
{"type":"Feature","geometry":{"type":"Point","coordinates":[-3.6,40.4]},"properties":{"name":"Start","brand":"Cepsa","address":"A-3 km 5","city":"Madrid","country":"ES","fuelType":"B7","currency":"EUR","price":1.60,"routeFraction":0.02}},
{"type":"Feature","geometry":{"type":"Point","coordinates":[-0.5,39.5]},"properties":{"name":"End","brand":null,"address":"V-30","city":"Valencia","country":"ES","fuelType":"B7","currency":"EUR","price":1.80,"routeFraction":0.98}}
]}`

func TestFindRouteStations_forwards_geometry_and_ranks(t *testing.T) {
	api := newFakeAPI(t, map[string]string{"POST /api/route-stations": routeStationsFixture})
	_, h := NewFindRouteStations(api.client())

	geom := map[string]any{"type": "LineString", "coordinates": []any{[]any{-3.7, 40.4}, []any{-0.4, 39.5}}}
	res := as[StationList](t, call(t, h, map[string]any{"geometry": geom, "fuel": "B7", "sort_by": "route"}))

	body := api.last(t, "/api/route-stations").Body
	if body["fuel"] != "B7" || body["corridorKm"] != 5.0 || fmt.Sprint(body["geometry"]) != "map[coordinates:[[-3.7 40.4] [-0.4 39.5]] type:LineString]" {
		t.Errorf("body = %v", body)
	}
	if res.Stations[0].Name != "Start" || res.Stations[2].Name != "End" {
		t.Errorf("not in route order: %+v", res.Stations)
	}
	if res.Stations[1].RouteKm == nil || *res.Stations[1].RouteKm < 100 || *res.Stations[1].RouteKm > 200 {
		t.Errorf("route_km = %v", res.Stations[1].RouteKm)
	}
}

func TestFindRouteStations_accepts_geometry_as_json_text(t *testing.T) {
	api := newFakeAPI(t, map[string]string{"POST /api/route-stations": routeStationsFixture})
	_, h := NewFindRouteStations(api.client())
	res := as[StationList](t, call(t, h, map[string]any{
		"geometry": `{"type":"LineString","coordinates":[[-3.7,40.4],[-0.4,39.5]]}`, "fuel": "B7",
	}))
	if res.Stations[0].Name != "Start" {
		t.Errorf("not sorted by price: %+v", res.Stations[0])
	}
}

func TestFindRouteStations_rejects_bad_geometry(t *testing.T) {
	api := newFakeAPI(t, map[string]string{"POST /api/route-stations": routeStationsFixture})
	_, h := NewFindRouteStations(api.client())
	for name, tc := range map[string]struct {
		geom any
		want string
	}{
		"missing":       {nil, "missing required parameter: geometry"},
		"not json":      {"{nope", "GeoJSON LineString object"},
		"point":         {map[string]any{"type": "Point", "coordinates": []any{1.0, 2.0}}, "GeoJSON LineString object"},
		"one point":     {map[string]any{"type": "LineString", "coordinates": []any{[]any{1.0, 2.0}}}, "2 to 2000 points"},
		"lat lon order": {map[string]any{"type": "LineString", "coordinates": []any{[]any{40.0, 100.0}, []any{1.0, 2.0}}}, "longitude first"},
		"polygon":       {map[string]any{"type": "Polygon", "coordinates": []any{}}, "type must be LineString"},
	} {
		t.Run(name, func(t *testing.T) {
			wantError(t, call(t, h, map[string]any{"geometry": tc.geom, "fuel": "B7"}), tc.want)
		})
	}
}

func TestDownsample_keeps_ends_and_limit(t *testing.T) {
	var pts [][]float64
	for i := 0; i < 4321; i++ {
		pts = append(pts, []float64{float64(i), 0})
	}
	got := downsample(pts, 2000)
	if len(got) != 2000 || got[0][0] != 0 || got[1999][0] != 4320 {
		t.Errorf("len %d, first %v, last %v", len(got), got[0], got[len(got)-1])
	}
	if short := downsample(pts[:10], 2000); len(short) != 10 {
		t.Errorf("short line changed: %d", len(short))
	}
}
