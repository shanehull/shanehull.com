package handlers

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestIncomeAsOf(t *testing.T) {
	points := []incomePoint{
		{Date: date(2020, 5, 1), Value: 1500},
		{Date: date(2020, 11, 1), Value: 1550},
	}
	if v, ok := incomeAsOf(points, date(2020, 8, 1)); !ok || v != 1500 {
		t.Fatalf("want forward-filled 1500, got %v %v", v, ok)
	}
	if _, ok := incomeAsOf(points, date(2020, 1, 1)); ok {
		t.Fatal("want no value before the first observation")
	}
}

func TestPriceToIncomeControls(t *testing.T) {
	rangeParam, showQuartiles := priceToIncomeControls(httptest.NewRequest("GET", "/house-prices-to-income/chart", nil))
	if rangeParam != "max" || showQuartiles {
		t.Fatalf("defaults = %q %v, want max false", rangeParam, showQuartiles)
	}
	rangeParam, showQuartiles = priceToIncomeControls(httptest.NewRequest("GET", "/house-prices-to-income/chart?range=20y&quartiles=on", nil))
	if rangeParam != "20y" || !showQuartiles {
		t.Fatalf("params = %q %v, want 20y true", rangeParam, showQuartiles)
	}
}

func TestPriceToIncomeSeriesRangeFilter(t *testing.T) {
	now := time.Now()
	points := []incomePoint{
		{Date: now.AddDate(-10, 0, 0), Value: 0.05},
		{Date: now.AddDate(-1, 0, 0), Value: 0.10},
	}
	if got := priceToIncomeSeries(points, "max", false); len(got) != 2 {
		t.Fatalf("max range: want 2 rows, got %d", len(got))
	}
	if got := priceToIncomeSeries(points, "5y", false); len(got) != 1 {
		t.Fatalf("5y range: want 1 row, got %d", len(got))
	}
}

func TestPriceToIncomeSeriesQuartiles(t *testing.T) {
	now := time.Now()
	points := []incomePoint{
		{Date: now.AddDate(-12, 0, 0), Value: 1},
		{Date: now.AddDate(-9, 0, 0), Value: 2},
		{Date: now.AddDate(-6, 0, 0), Value: 3},
		{Date: now.AddDate(-3, 0, 0), Value: 4},
	}
	data := priceToIncomeSeries(points, "max", true)
	wantQ1 := 1.75
	wantQ3 := 3.25
	for i, row := range data {
		if row.Quartile1 != wantQ1 || row.Quartile3 != wantQ3 {
			t.Fatalf("row %d quartiles = %v/%v, want %v/%v", i, row.Quartile1, row.Quartile3, wantQ1, wantQ3)
		}
	}
}
