package oecd

import (
	"context"
	"net/http"
	"net/http/httptest"
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

func TestFetchPriceToRent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/OECD.ECO.MPD,DSD_AN_HOUSE_PRICES@DF_HOUSE_PRICES,1.0/AUS.Q.HPI_RPI." {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte("REF_AREA,TIME_PERIOD,OBS_VALUE\n" +
			"AUS,2025-Q1,130\n" +
			"AUS,2025-Q2,131\n"))
	}))
	defer server.Close()

	points, err := New(WithBaseURL(server.URL)).FetchPriceToRent(context.Background(), "AUS")
	if err != nil {
		t.Fatal(err)
	}
	if len(points) != 2 || points[1].Value != 131 {
		t.Fatalf("unexpected points: %+v", points)
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
