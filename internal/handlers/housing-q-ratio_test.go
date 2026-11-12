package handlers

import (
	"math"
	"net/http/httptest"
	"testing"
	"time"
)

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func TestQRatio(t *testing.T) {
	equity := []qPoint{
		{Date: date(2020, 3, 1), Value: 100},
		{Date: date(2020, 6, 1), Value: 120},
	}
	replacement := []qPoint{
		{Date: date(2020, 3, 1), Value: 50},
		{Date: date(2020, 6, 1), Value: 60},
	}
	got := qRatio(equity, replacement)
	if len(got) != 2 || got[0].Value != 2 || got[1].Value != 2 {
		t.Fatalf("unexpected ratio: %+v", got)
	}
}

func TestQAsOf(t *testing.T) {
	points := []qPoint{
		{Date: date(2020, 3, 1), Value: 10},
		{Date: date(2020, 6, 1), Value: 20},
	}
	if v, ok := qAsOf(points, date(2020, 5, 1)); !ok || v != 10 {
		t.Fatalf("want forward-filled 10, got %v %v", v, ok)
	}
	if _, ok := qAsOf(points, date(2020, 1, 1)); ok {
		t.Fatal("want no value before the first observation")
	}
}

func TestQDetrend(t *testing.T) {
	points := []qPoint{
		{Date: date(2020, 3, 1), Value: 1},
		{Date: date(2020, 6, 1), Value: 2},
		{Date: date(2020, 9, 1), Value: 4},
	}
	got := qDetrend(points)
	want := []float64{1, math.Sqrt2, 2}
	for i := range want {
		if math.Abs(got[i].Value-want[i]) > 1e-9 {
			t.Fatalf("qDetrend[%d] = %v, want %v", i, got[i].Value, want[i])
		}
	}
}

func TestQRatioControls(t *testing.T) {
	rangeParam, detrended, showQuartiles := qRatioControls(httptest.NewRequest("GET", "/housing-q-ratio/chart", nil))
	if rangeParam != "max" || !detrended || showQuartiles {
		t.Fatalf("defaults = %q %v %v, want max true false", rangeParam, detrended, showQuartiles)
	}

	req := httptest.NewRequest("GET", "/housing-q-ratio/chart?range=10y&detrend=off&quartiles=on", nil)
	rangeParam, detrended, showQuartiles = qRatioControls(req)
	if rangeParam != "10y" || detrended || !showQuartiles {
		t.Fatalf("params = %q %v %v, want 10y false true", rangeParam, detrended, showQuartiles)
	}

	req = httptest.NewRequest("GET", "/housing-q-ratio/chart?detrend=on", nil)
	_, detrended, _ = qRatioControls(req)
	if !detrended {
		t.Fatal("detrend=on should keep detrending")
	}

	// The checked box and its off fallback both post; order must not matter.
	for _, query := range []string{"detrend=on&detrend=off", "detrend=off&detrend=on"} {
		req = httptest.NewRequest("GET", "/housing-q-ratio/chart?"+query, nil)
		_, detrended, _ = qRatioControls(req)
		if !detrended {
			t.Fatalf("%s should keep detrending", query)
		}
	}
}

func TestQRatioSeriesRangeFilter(t *testing.T) {
	now := time.Now()
	points := []qPoint{
		{Date: now.AddDate(-10, 0, 0), Value: 1},
		{Date: now.AddDate(-1, 0, 0), Value: 2},
	}
	if got := qRatioSeries(points, "max", false, false); len(got) != 2 {
		t.Fatalf("max range: want 2 rows, got %d", len(got))
	}
	if got := qRatioSeries(points, "5y", false, false); len(got) != 1 {
		t.Fatalf("5y range: want 1 row, got %d", len(got))
	}
}

func TestQRatioSeriesDetrendsFullHistory(t *testing.T) {
	now := time.Now()
	points := []qPoint{
		{Date: now.AddDate(-12, 0, 0), Value: 1},
		{Date: now.AddDate(-9, 0, 0), Value: 2},
		{Date: now.AddDate(-6, 0, 0), Value: 3},
		{Date: now.AddDate(-3, 0, 0), Value: 4},
	}
	full := qRatioSeries(points, "max", true, false)
	short := qRatioSeries(points, "5y", true, false)
	if len(short) == 0 {
		t.Fatal("want rows in the short range")
	}
	last := short[len(short)-1]
	if last.Value != full[len(full)-1].Value {
		t.Fatalf("detrended value changed with range: %v vs %v", last.Value, full[len(full)-1].Value)
	}
}

func TestQRatioSeriesQuartiles(t *testing.T) {
	now := time.Now()
	points := []qPoint{
		{Date: now.AddDate(-12, 0, 0), Value: 1},
		{Date: now.AddDate(-9, 0, 0), Value: 2},
		{Date: now.AddDate(-6, 0, 0), Value: 3},
		{Date: now.AddDate(-3, 0, 0), Value: 4},
	}
	data := qRatioSeries(points, "max", false, true)
	if len(data) != 4 {
		t.Fatalf("want 4 rows, got %d", len(data))
	}
	// Quartiles are constant across the series and bracket the middle values.
	wantQ1 := 1.75
	wantQ3 := 3.25
	for i, row := range data {
		if row.Quartile1 != wantQ1 || row.Quartile3 != wantQ3 {
			t.Fatalf("row %d quartiles = %v/%v, want %v/%v", i, row.Quartile1, row.Quartile3, wantQ1, wantQ3)
		}
	}

	// Without the flag the rows carry no quartiles.
	for _, row := range qRatioSeries(points, "max", false, false) {
		if row.Quartile1 != 0 || row.Quartile3 != 0 {
			t.Fatalf("unexpected quartiles when disabled: %+v", row)
		}
	}
}
