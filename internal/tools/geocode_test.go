package tools

import "testing"

func TestGeocode_returns_places_lat_lon(t *testing.T) {
	api := newFakeAPI(t, map[string]string{"GET /api/geocode": geocodeFixture})
	_, h := NewGeocode(api.client())

	res := as[GeocodeResult](t, call(t, h, map[string]any{"query": " Atocha ", "lat": 40.4, "lon": -3.7, "limit": 1}))

	q := api.last(t, "/api/geocode").Query
	if q["q"] != "Atocha" || q["lat"] != "40.4" || q["lon"] != "-3.7" {
		t.Errorf("query = %v", q)
	}
	if len(res.Results) != 1 {
		t.Fatalf("results = %+v", res.Results)
	}
	p := res.Results[0]
	if p.Lat != 40.406975 || p.Lon != -3.688582 || p.City != "Madrid" || p.State != "Comunidad de Madrid" {
		t.Errorf("place = %+v", p)
	}
}

func TestGeocode_null_fields_are_omitted(t *testing.T) {
	api := newFakeAPI(t, map[string]string{"GET /api/geocode": geocodeFixture})
	_, h := NewGeocode(api.client())
	res := as[GeocodeResult](t, call(t, h, map[string]any{"query": "Atocha"}))
	if len(res.Results) != 2 || res.Results[1].State != "" {
		t.Errorf("results = %+v", res.Results)
	}
	if _, ok := api.last(t, "/api/geocode").Query["lat"]; ok {
		t.Error("lat sent without a bias point")
	}
}

func TestGeocode_rejects_bad_input(t *testing.T) {
	api := newFakeAPI(t, nil)
	_, h := NewGeocode(api.client())
	long := make([]rune, 201)
	for i := range long {
		long[i] = 'a'
	}
	for name, tc := range map[string]struct {
		args map[string]any
		want string
	}{
		"no query":    {map[string]any{}, "query"},
		"blank query": {map[string]any{"query": "  "}, "1 to 200 characters"},
		"long query":  {map[string]any{"query": string(long)}, "1 to 200 characters"},
		"lat only":    {map[string]any{"query": "x", "lat": 1.0}, "both lat and lon"},
		"limit 0":     {map[string]any{"query": "x", "limit": 0}, "limit must be between"},
	} {
		t.Run(name, func(t *testing.T) { wantError(t, call(t, h, tc.args), tc.want) })
	}
	if n := len(api.requests()); n != 0 {
		t.Errorf("%d upstream requests for invalid input", n)
	}
}
