package wb

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const obsJSON = `[{"date":"2023","value":24.4},{"date":"2022","value":24.6},{"date":"2021","value":null}]`

func newClient(handler http.Handler) *Client {
	srv := httptest.NewServer(handler)
	return New(WithBaseURL(srv.URL))
}

func TestFetchIndicator(t *testing.T) {
	client := newClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/country/USA/indicator/GC.XPN.TOTL.GD.ZS" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		_, _ = fmt.Fprintf(w, `[{"page":1,"total":66},%s]`, obsJSON)
	}))

	points, err := client.FetchIndicator("USA", "GC.XPN.TOTL.GD.ZS")
	if err != nil {
		t.Fatalf("FetchIndicator: %v", err)
	}

	// Null 2021 dropped, sorted oldest first.
	if len(points) != 2 {
		t.Fatalf("got %d points, want 2", len(points))
	}
	if !points[0].Date.Equal(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)) || points[0].Value != 24.6 {
		t.Errorf("unexpected first point: %v %v", points[0].Date, points[0].Value)
	}
	if points[1].Value != 24.4 {
		t.Errorf("unexpected last value: %v", points[1].Value)
	}
}

func TestFetchIndicatorEmpty(t *testing.T) {
	client := newClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `[{"page":1,"total":0},[]]`)
	}))
	if _, err := client.FetchIndicator("USA", "X"); err == nil {
		t.Error("expected error for empty observations")
	}
}

func TestFetchIndicatorNoRows(t *testing.T) {
	client := newClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `[{"page":1,"total":3},[{"date":"2020","value":null}]]`)
	}))
	if _, err := client.FetchIndicator("USA", "X"); err == nil {
		t.Error("expected error when every value is null")
	}
}

func TestFetchIndicatorBadStatus(t *testing.T) {
	client := newClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	if _, err := client.FetchIndicator("USA", "X"); err == nil {
		t.Error("expected error for non-200 status")
	}
}

func TestFetchIndicatorMalformed(t *testing.T) {
	client := newClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `not json`)
	}))
	if _, err := client.FetchIndicator("USA", "X"); err == nil {
		t.Error("expected parse error")
	}
}
