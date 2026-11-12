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

func TestWeeklyIncomeControls(t *testing.T) {
	if got := weeklyIncomeControls(httptest.NewRequest("GET", "/house-prices-to-weekly-income/chart", nil)); got != "max" {
		t.Fatalf("default range = %q, want max", got)
	}
	if got := weeklyIncomeControls(httptest.NewRequest("GET", "/house-prices-to-weekly-income/chart?range=20y", nil)); got != "20y" {
		t.Fatalf("range = %q, want 20y", got)
	}
}

func TestWeeklyIncomeSeriesRangeFilter(t *testing.T) {
	now := time.Now()
	points := []incomePoint{
		{Date: now.AddDate(-10, 0, 0), Value: 0.05},
		{Date: now.AddDate(-1, 0, 0), Value: 0.10},
	}
	if got := weeklyIncomeSeries(points, "max"); len(got) != 2 {
		t.Fatalf("max range: want 2 rows, got %d", len(got))
	}
	if got := weeklyIncomeSeries(points, "5y"); len(got) != 1 {
		t.Fatalf("5y range: want 1 row, got %d", len(got))
	}
}
