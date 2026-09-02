// Package yahoo is a minimal client for the public Yahoo Finance chart API.
// It returns monthly (or other interval) settlement prices for any symbol,
// e.g. COMEX gold futures (GC=F). The client is stateless so it can be shared
// across tools; callers apply their own caching.
package yahoo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// PricePoint is a settlement price observed at a point in time.
type PricePoint struct {
	Date  time.Time
	Price float64
}

const defaultBaseURL = "https://query1.finance.yahoo.com/v8/finance/chart"

// Client fetches chart data from the Yahoo Finance API.
type Client struct {
	baseURL    string
	httpClient *http.Client
	userAgent  string
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
		userAgent:  "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// FetchMonthly returns a symbol's monthly settlement prices from the earliest
// available month to today, oldest first. Months with a null price are omitted.
func (c *Client) FetchMonthly(symbol string) ([]PricePoint, error) {
	return c.fetch(symbol, "1mo", "max")
}

func (c *Client) fetch(symbol, interval, rng string) ([]PricePoint, error) {
	endpoint := strings.TrimRight(c.baseURL, "/") + "/" + url.PathEscape(symbol)
	query := url.Values{}
	query.Set("range", rng)
	query.Set("interval", interval)

	req, err := http.NewRequest(http.MethodGet, endpoint+"?"+query.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", symbol, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch %s: unexpected status %d", symbol, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read %s response: %w", symbol, err)
	}

	var parsed chartResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("parse %s response: %w", symbol, err)
	}
	if parsed.Chart.Error != nil {
		return nil, fmt.Errorf("fetch %s: API error %s", symbol, strings.TrimSpace(string(*parsed.Chart.Error)))
	}
	if len(parsed.Chart.Result) == 0 {
		return nil, fmt.Errorf("fetch %s: no data", symbol)
	}

	series := parsed.Chart.Result[0]
	if len(series.Indicators.Quote) == 0 {
		return nil, fmt.Errorf("fetch %s: no quote data", symbol)
	}
	closes := series.Indicators.Quote[0].Close

	points := make([]PricePoint, 0, len(series.Timestamp))
	for i, ts := range series.Timestamp {
		if i >= len(closes) || closes[i] == nil {
			continue
		}
		points = append(points, PricePoint{
			Date:  time.Unix(ts, 0).UTC(),
			Price: *closes[i],
		})
	}

	if len(points) == 0 {
		return nil, fmt.Errorf("fetch %s: no prices", symbol)
	}

	return points, nil
}

type chartResponse struct {
	Chart chart `json:"chart"`
}

type chart struct {
	Result []chartData      `json:"result"`
	Error  *json.RawMessage `json:"error"`
}

type chartData struct {
	Timestamp  []int64 `json:"timestamp"`
	Indicators struct {
		Quote []struct {
			Close []*float64 `json:"close"`
		} `json:"quote"`
	} `json:"indicators"`
}
