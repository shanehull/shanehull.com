package handlers

import (
	"testing"
	"time"

	"github.com/shanehull/shanehull.com/internal/cfs"
)

func TestMergeM4Gold(t *testing.T) {
	m4 := []cfs.Point{
		{Date: time.Date(1967, 1, 1, 0, 0, 0, 0, time.UTC), Index: 100},
		{Date: time.Date(1971, 1, 1, 0, 0, 0, 0, time.UTC), Index: 150},
		{Date: time.Date(1971, 3, 1, 0, 0, 0, 0, time.UTC), Index: 160},
	}
	gold := []goldPricePoint{
		{Date: time.Date(1967, 1, 1, 0, 0, 0, 0, time.UTC), Price: 35.09},
		{Date: time.Date(1971, 1, 1, 0, 0, 0, 0, time.UTC), Price: 43.00},
		{Date: time.Date(1971, 2, 1, 0, 0, 0, 0, time.UTC), Price: 40.00},
	}

	points := mergeM4Gold(m4, gold)
	if len(points) != 2 {
		t.Fatalf("got %d points, want 2 (only months both series share)", len(points))
	}

	// 1967: 100 x 35.09 / 35.09 = 100 (the base).
	if points[0].Index < 99.9 || points[0].Index > 100.1 {
		t.Errorf("1967 index = %v, want 100", points[0].Index)
	}
	// 1971-01: 150 x 35.09 / 43 = 122.4.
	want := 150 * 35.09 / 43
	if points[1].Index < want-0.01 || points[1].Index > want+0.01 {
		t.Errorf("1971 index = %v, want %v", points[1].Index, want)
	}
}

func TestMergeM4GoldNoOverlap(t *testing.T) {
	m4 := []cfs.Point{{Date: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), Index: 1500}}
	gold := []goldPricePoint{{Date: time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC), Price: 600}}
	if points := mergeM4Gold(m4, gold); len(points) != 0 {
		t.Errorf("expected no overlap, got %d points", len(points))
	}
}