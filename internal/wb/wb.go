// Package wb is a minimal client for the World Bank indicator API. It returns
// one WDI indicator's annual values for a country, e.g. general government
// expense as a share of GDP. The client is stateless; callers apply their own
// caching.
package wb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"
)

// Point is an annual observation for a country.
type Point struct {
	Date  time.Time
	Value float64
}

const defaultBaseURL = "https://api.worldbank.org/v2"

// Client fetches indicators from the World Bank API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// Option configures a Client.
type Option func(*Client)

// WithBaseURL overrides the API endpoint. Used by tests.
func WithBaseURL(url string) Option {
	return func(c *Client) { c.baseURL = url }
}

// WithHTTPClient overrides the HTTP client, including its timeout.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) { c.httpClient = client }
}

// New returns a Client with sensible defaults.
func New(opts ...Option) *Client {
	c := &Client{
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// FetchIndicator returns an indicator's annual values for a country, oldest
// first. Years with missing values are omitted.
func (c *Client) FetchIndicator(country, indicator string) ([]Point, error) {
	endpoint := fmt.Sprintf("%s/country/%s/indicator/%s?format=json&per_page=1000&date=1960:2026",
		c.baseURL, url.PathEscape(country), url.PathEscape(indicator))

	resp, err := c.httpClient.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("fetch World Bank %s.%s: %w", country, indicator, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch World Bank %s.%s: status %d", country, indicator, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read World Bank response: %w", err)
	}

	var raw []json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parse World Bank response: %w", err)
	}
	if len(raw) < 2 {
		return nil, fmt.Errorf("no data from World Bank for %s.%s", country, indicator)
	}

	var obs []struct {
		Date  string   `json:"date"`
		Value *float64 `json:"value"`
	}
	if err := json.Unmarshal(raw[1], &obs); err != nil {
		return nil, fmt.Errorf("parse World Bank observations: %w", err)
	}

	points := make([]Point, 0, len(obs))
	for _, o := range obs {
		if o.Value == nil {
			continue
		}
		year, err := strconv.Atoi(o.Date)
		if err != nil {
			continue
		}
		points = append(points, Point{Date: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Value: *o.Value})
	}

	if len(points) == 0 {
		return nil, fmt.Errorf("no data from World Bank for %s.%s", country, indicator)
	}

	sort.Slice(points, func(i, j int) bool { return points[i].Date.Before(points[j].Date) })
	return points, nil
}
