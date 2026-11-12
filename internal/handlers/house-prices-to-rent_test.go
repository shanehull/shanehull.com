package handlers

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestPriceToRentControls(t *testing.T) {
	rangeParam, showQuartiles := priceToRentControls(httptest.NewRequest("GET", "/house-prices-to-rent/chart", nil))
	if rangeParam != "max" || showQuartiles {
		t.Fatalf("defaults = %q %v, want max false", rangeParam, showQuartiles)
	}
	rangeParam, showQuartiles = priceToRentControls(httptest.NewRequest("GET", "/house-prices-to-rent/chart?range=20y&quartiles=on", nil))
	if rangeParam != "20y" || !showQuartiles {
		t.Fatalf("params = %q %v, want 20y true", rangeParam, showQuartiles)
	}
}

func TestPriceToRentSeriesRangeFilter(t *testing.T) {
	now := time.Now()
	points := []rentPoint{
		{Date: now.AddDate(-10, 0, 0), Value: 80},
		{Date: now.AddDate(-1, 0, 0), Value: 130},
	}
	if got := priceToRentSeries(points, "max", false); len(got) != 2 {
		t.Fatalf("max range: want 2 rows, got %d", len(got))
	}
	if got := priceToRentSeries(points, "5y", false); len(got) != 1 {
		t.Fatalf("5y range: want 1 row, got %d", len(got))
	}
}

func TestPriceToRentSeriesQuartiles(t *testing.T) {
	now := time.Now()
	points := []rentPoint{
		{Date: now.AddDate(-12, 0, 0), Value: 1},
		{Date: now.AddDate(-9, 0, 0), Value: 2},
		{Date: now.AddDate(-6, 0, 0), Value: 3},
		{Date: now.AddDate(-3, 0, 0), Value: 4},
	}
	data := priceToRentSeries(points, "max", true)
	wantQ1 := 1.75
	wantQ3 := 3.25
	for i, row := range data {
		if row.Quartile1 != wantQ1 || row.Quartile3 != wantQ3 {
			t.Fatalf("row %d quartiles = %v/%v, want %v/%v", i, row.Quartile1, row.Quartile3, wantQ1, wantQ3)
		}
	}
}
