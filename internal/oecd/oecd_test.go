package oecd

import (
	"testing"
	"time"
)

func TestParsePeriod(t *testing.T) {
	cases := map[string]time.Time{
		"2026-Q2": time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		"1970-Q3": time.Date(1970, 7, 1, 0, 0, 0, 0, time.UTC),
	}
	for in, want := range cases {
		got, err := parsePeriod(in)
		if err != nil {
			t.Fatalf("parsePeriod(%q): %v", in, err)
		}
		if !got.Equal(want) {
			t.Fatalf("parsePeriod(%q) = %v, want %v", in, got, want)
		}
	}
	if _, err := parsePeriod("nonsense"); err == nil {
		t.Fatal("expected an error for a nonsense period")
	}
}

func TestParsePeriodValueCSV(t *testing.T) {
	body := []byte("REF_AREA,TIME_PERIOD,OBS_VALUE\n" +
		"AUS,1970-Q3,2.58\n" +
		"AUS,1970-Q4,3.12\n" +
		"AUS,1971-Q1,\n")
	points, err := parsePeriodValueCSV(body)
	if err != nil {
		t.Fatal(err)
	}
	if len(points) != 2 {
		t.Fatalf("want 2 points, got %d", len(points))
	}
	if points[0].Value != 2.58 || points[1].Value != 3.12 {
		t.Fatalf("unexpected values: %+v", points)
	}
}
