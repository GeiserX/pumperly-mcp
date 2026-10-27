package tools

import (
	"strings"
	"testing"
)

func TestGetStationsInArea_sorts_by_price_and_limits(t *testing.T) {
	api := newFakeAPI(t, map[string]string{"GET /api/stations": nearestFixture})
	_, h := NewGetStationsInArea(api.client())

	res := as[StationList](t, call(t, h, map[string]any{"bbox": " -3.75, 40.38,-3.65,40.45", "fuel": "B7", "limit": 2}))

	q := api.last(t, "/api/stations").Query
	if q["bbox"] != "-3.75,40.38,-3.65,40.45" || q["fuel"] != "B7" {
		t.Errorf("query = %v", q)
	}
	if res.SortBy != "price" || res.Found != 4 || res.Returned != 2 || res.Stations[0].Price == nil || *res.Stations[0].Price != 1.659 {
		t.Errorf("list = %+v", res)
	}
	for _, s := range res.Stations {
		if s.DistanceKm == nil {
			t.Errorf("%s has no distance from the box centre", s.Name)
		}
	}
}

func TestGetStationsInArea_unpriced_stations_sort_last(t *testing.T) {
	api := newFakeAPI(t, map[string]string{"GET /api/stations": nearestFixture})
	_, h := NewGetStationsInArea(api.client())
	res := as[StationList](t, call(t, h, map[string]any{"bbox": "-3.75,40.38,-3.65,40.45", "fuel": "B7"}))
	if last := res.Stations[len(res.Stations)-1]; last.Name != "No Price" || last.Price != nil || last.Currency != "" {
		t.Errorf("last = %+v", last)
	}
}

func TestParseBBox(t *testing.T) {
	for in, want := range map[string]string{
		"-3.75,40.38,-3.65,40.45": "",
		"-3.65,40.38,-3.75,40.45": "min < max",
		"1,2,3":                   "four numbers",
		"a,2,3,4":                 "not a number",
		"-200,40,-3,41":           "longitude must be between",
	} {
		_, err := parseBBox(in)
		switch {
		case want == "" && err != nil:
			t.Errorf("parseBBox(%q) = %v", in, err)
		case want != "" && (err == nil || !strings.Contains(err.Error(), want)):
			t.Errorf("parseBBox(%q) = %v, want %q", in, err, want)
		}
	}
}
