package yahoo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func chartJSON(timestamps []int64, closes []*float64) string {
	ts, _ := json.Marshal(timestamps)
	cs, _ := json.Marshal(closes)
	return `{"chart":{"result":[{"meta":{},"timestamp":` + string(ts) + `,"indicators":{"quote":[{"close":` + string(cs) + `}]}}],"error":null}}`
}

func f(v float64) *float64 { return &v }

func epoch(year int) int64 {
	return time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
}

func newTestServer(handler http.HandlerFunc) *Client {
	srv := httptest.NewServer(handler)
	return New(WithBaseURL(srv.URL))
}

func TestFetchMonthlyOrderAndSkipNulls(t *testing.T) {
	ts := []int64{epoch(2000), epoch(2001), epoch(2002)}
	closes := []*float64{f(300), nil, f(415.3)}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/GC=F" && r.URL.Path != "/GC%3DF" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("interval"); got != "1mo" {
			t.Errorf("interval = %q, want 1mo", got)
		}
		if got := r.URL.Query().Get("range"); got != "max" {
			t.Errorf("range = %q, want max", got)
		}
		fmt.Fprint(w, chartJSON(ts, closes))
	}))
	defer srv.Close()

	points, err := New(WithBaseURL(srv.URL)).FetchMonthly("GC=F")
	if err != nil {
		t.Fatalf("FetchMonthly: %v", err)
	}

	wantDates := []time.Time{time.Unix(ts[0], 0).UTC(), time.Unix(ts[2], 0).UTC()}
	if len(points) != 2 {
		t.Fatalf("got %d points, want 2 (null month skipped)", len(points))
	}
	for i, p := range points {
		if !p.Date.Equal(wantDates[i]) {
			t.Errorf("point %d date = %v, want %v", i, p.Date, wantDates[i])
		}
	}
	if points[0].Price != 300 {
		t.Errorf("first price = %v, want 300", points[0].Price)
	}
	if points[1].Price != 415.3 {
		t.Errorf("second price = %v, want 415.3", points[1].Price)
	}
}

func TestFetchMonthlyAPISetsErrorField(t *testing.T) {
	client := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"chart":{"result":[],"error":{"code":"Not Found"}}}`)
	}))
	if _, err := client.FetchMonthly("GC=F"); err == nil || !strings.Contains(err.Error(), "Not Found") {
		t.Errorf("expected API error mentioning Not Found, got %v", err)
	}
}

func TestFetchMonthlyBadStatus(t *testing.T) {
	client := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "rate limited", http.StatusTooManyRequests)
	}))
	if _, err := client.FetchMonthly("GC=F"); err == nil || !strings.Contains(err.Error(), "429") {
		t.Errorf("expected 429 error, got %v", err)
	}
}

func TestFetchMonthlyEmptyResult(t *testing.T) {
	client := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"chart":{"result":[],"error":null}}`)
	}))
	if _, err := client.FetchMonthly("GC=F"); err == nil {
		t.Error("expected error for empty result")
	}
}

func TestFetchMonthlyNoQuoteData(t *testing.T) {
	client := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"chart":{"result":[{"timestamp":[1],"indicators":{"quote":[]}}],"error":null}}`)
	}))
	if _, err := client.FetchMonthly("GC=F"); err == nil {
		t.Error("expected error for missing quote data")
	}
}

func TestFetchMonthlyNoPrices(t *testing.T) {
	client := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"chart":{"result":[{"timestamp":[1],"indicators":{"quote":[{"close":[null]}]}}],"error":null}}`)
	}))
	if _, err := client.FetchMonthly("GC=F"); err == nil {
		t.Error("expected error when every price is null")
	}
}

func TestFetchMonthlyMalformedBody(t *testing.T) {
	client := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `not json at all`)
	}))
	if _, err := client.FetchMonthly("GC=F"); err == nil {
		t.Error("expected parse error")
	}
}

func TestFetchMonthlyNetworkError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := srv.URL
	srv.Close()

	client := New(WithBaseURL(url))
	if _, err := client.FetchMonthly("GC=F"); err == nil {
		t.Error("expected network error")
	}
}

func TestFetchMonthlyEscapesSymbol(t *testing.T) {
	var gotPath string
	client := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		fmt.Fprint(w, `{"chart":{"result":[{"timestamp":[1],"indicators":{"quote":[{"close":[50]}]}}],"error":null}}`)
	}))
	if _, err := client.FetchMonthly("GC=F"); err != nil {
		t.Fatalf("FetchMonthly: %v", err)
	}

	escaped := url.PathEscape("GC=F")
	if gotPath != "/"+escaped && gotPath != "/GC=F" {
		t.Errorf("got symbol path %q", gotPath)
	}
}

func TestWithHTTPClientApplies(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"chart":{"result":[{"timestamp":[1],"indicators":{"quote":[{"close":[50]}]}}],"error":null}}`)
	}))
	defer srv.Close()

	custom := &http.Client{Timeout: 5 * time.Second}
	client := New(WithBaseURL(srv.URL), WithHTTPClient(custom))
	if client.httpClient != custom {
		t.Error("WithHTTPClient did not apply")
	}
	if _, err := client.FetchMonthly("GC=F"); err != nil {
		t.Fatalf("FetchMonthly with custom client: %v", err)
	}
}

func TestDefaultClientBaseURL(t *testing.T) {
	if got := New().baseURL; got != defaultBaseURL {
		t.Errorf("default base URL = %q, want %q", got, defaultBaseURL)
	}
}
