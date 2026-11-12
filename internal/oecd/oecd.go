// Package oecd fetches the Organisation for Economic Co-operation and
// Development nominal house price index. The client is stateless; callers apply
// their own caching.
package oecd

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Point is one quarterly observation.
type Point struct {
	Date  time.Time
	Value float64
}

const defaultBaseURL = "https://sdmx.oecd.org/public/rest/data"

// Client fetches OECD data series.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// Option configures a Client.
type Option func(*Client)

// WithBaseURL overrides the API endpoint. Used by tests.
func WithBaseURL(rawURL string) Option {
	return func(c *Client) { c.baseURL = rawURL }
}

// WithHTTPClient overrides the HTTP client, including its timeout.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) { c.httpClient = client }
}

// New returns a Client with sensible defaults.
func New(opts ...Option) *Client {
	c := &Client{
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// FetchHousePriceIndex returns the nominal house price index for a country
// (2015 = 100), quarterly and oldest first.
func (c *Client) FetchHousePriceIndex(ctx context.Context, country string) ([]Point, error) {
	endpoint := c.baseURL + "/OECD.ECO.MPD,DSD_AN_HOUSE_PRICES@DF_HOUSE_PRICES,1.0/" +
		url.PathEscape(country) + ".Q.HPI.?format=csvfile"

	body, err := c.get(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	points, err := parsePeriodValueCSV(body)
	if err != nil {
		return nil, fmt.Errorf("parse OECD house price index: %w", err)
	}
	if len(points) == 0 {
		return nil, fmt.Errorf("no OECD house price observations for %s", country)
	}
	return points, nil
}

func (c *Client) get(ctx context.Context, endpoint string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
	req.Header.Set("Accept", "text/csv")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch OECD: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch OECD: status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read OECD response: %w", err)
	}
	return body, nil
}

// parsePeriodValueCSV reads a CSV with TIME_PERIOD and OBS_VALUE columns.
func parsePeriodValueCSV(body []byte) ([]Point, error) {
	r := csv.NewReader(strings.NewReader(string(body)))
	r.FieldsPerRecord = -1
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("no rows")
	}

	header := rows[0]
	periodCol, valueCol := -1, -1
	for i, h := range header {
		switch strings.TrimSpace(h) {
		case "TIME_PERIOD":
			periodCol = i
		case "OBS_VALUE":
			valueCol = i
		}
	}
	if periodCol < 0 || valueCol < 0 {
		return nil, fmt.Errorf("unexpected header: %v", header)
	}

	var points []Point
	for _, row := range rows[1:] {
		if periodCol >= len(row) || valueCol >= len(row) {
			continue
		}
		t, err := parsePeriod(row[periodCol])
		if err != nil {
			continue
		}
		v, err := strconv.ParseFloat(strings.TrimSpace(row[valueCol]), 64)
		if err != nil {
			continue
		}
		points = append(points, Point{Date: t, Value: v})
	}
	sort.Slice(points, func(i, j int) bool { return points[i].Date.Before(points[j].Date) })
	return points, nil
}

// parsePeriod converts an OECD period string such as "2026-Q2" into the first
// day of the quarter.
func parsePeriod(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	i := strings.IndexAny(s, "Qq")
	if i < 0 {
		return time.Time{}, fmt.Errorf("unrecognised period %q", s)
	}
	year, err := strconv.Atoi(strings.TrimRight(strings.TrimSpace(s[:i]), "-_ "))
	if err != nil {
		return time.Time{}, err
	}
	quarter, err := strconv.Atoi(strings.TrimSpace(s[i+1:]))
	if err != nil || quarter < 1 || quarter > 4 {
		return time.Time{}, fmt.Errorf("unrecognised quarter %q", s)
	}
	return time.Date(year, time.Month((quarter-1)*3+1), 1, 0, 0, 0, 0, time.UTC), nil
}
