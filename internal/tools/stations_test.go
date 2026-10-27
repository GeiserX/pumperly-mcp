package tools

import "testing"

func f(v float64) *float64 { return &v }

func TestDecodeStations_maps_the_real_shape(t *testing.T) {
	st, err := decodeStations([]byte(nearestFixture), "B7", 0)
	if err != nil {
		t.Fatal(err)
	}
	s := st[0]
	if s.Name != "Blanca Madrid" || s.Brand != "Blanca" || s.Country != "ES" || s.Lat != 40.405278 || s.Lon != -3.703139 ||
		*s.Price != 1.879 || s.Currency != "EUR" || s.Unit != "L" || s.ReportedAt != "2026-10-27T04:37:35.719Z" || *s.DistanceKm != 1.29 {
		t.Errorf("station = %+v", s)
	}
	if st[2].Brand != "" {
		t.Errorf("null brand = %q", st[2].Brand)
	}
}

func TestDecodeStations_ev_fields(t *testing.T) {
	body := `{"type":"FeatureCollection","features":[{"type":"Feature","geometry":{"type":"Point","coordinates":[2,41]},"properties":{"name":"Charger","brand":null,"address":"","city":"Barcelona","fuelType":"EV","currency":"EUR","price":0.49,"powerKw":150,"evFees":{"sessionFee":1}}}]}`
	st, err := decodeStations([]byte(body), "EV", 0)
	if err != nil {
		t.Fatal(err)
	}
	if st[0].Unit != "kWh" || *st[0].PowerKw != 150 || st[0].EVFees["sessionFee"] != 1.0 {
		t.Errorf("station = %+v", st[0])
	}
}

func TestSummarisePrices_per_currency(t *testing.T) {
	got := summarisePrices([]Station{
		{Price: f(1.5), Currency: "EUR"}, {Price: f(1.7), Currency: "EUR"}, {Price: f(1.6), Currency: "EUR"}, {Price: f(1.9), Currency: "EUR"},
		{Price: f(7.1), Currency: "PLN"},
		{Price: nil},
	})
	if len(got) != 2 {
		t.Fatalf("summaries = %+v", got)
	}
	if e := got[0]; e.Currency != "EUR" || e.Count != 4 || e.Min != 1.5 || e.Max != 1.9 || e.Median != 1.65 {
		t.Errorf("EUR = %+v", e)
	}
	if p := got[1]; p.Currency != "PLN" || p.Median != 7.1 {
		t.Errorf("PLN = %+v", p)
	}
}

func TestNewStationList_sorts_then_limits(t *testing.T) {
	l := newStationList([]Station{
		{Name: "b", Price: f(2)}, {Name: "none"}, {Name: "a", Price: f(1)}, {Name: "c", Price: f(3)},
	}, "E5", "price", 3)
	if l.Found != 4 || l.Returned != 3 || l.Stations[0].Name != "a" || l.Stations[2].Name != "c" || l.Unit != "L" {
		t.Errorf("list = %+v", l)
	}
}
