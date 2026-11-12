// Package rba fetches Reserve Bank of Australia statistics. The client is
// stateless; callers apply their own caching.
package rba

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Point is one observation.
type Point struct {
	Date  time.Time
	Value float64
}

const (
	defaultBaseURL = "https://www.rba.gov.au"
	housingCredit  = "/statistics/tables/csv/d2-data.csv"
)

// Client fetches RBA data.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// Option configures a Client.
type Option func(*Client)

// WithBaseURL overrides the statistics endpoint. Used by tests.
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

// FetchHousingCredit returns housing credit outstanding, owner-occupier plus
// investor, in billions of dollars, monthly and oldest first.
func (c *Client) FetchHousingCredit(ctx context.Context) ([]Point, error) {
	body, err := c.get(ctx, c.baseURL+housingCredit)
	if err != nil {
		return nil, err
	}
	owner, err := parseColumn(body, "Credit; Owner-occupier housing")
	if err != nil {
		return nil, fmt.Errorf("parse RBA owner-occupier housing credit: %w", err)
	}
	investor, err := parseColumn(body, "Credit; Investor housing")
	if err != nil {
		return nil, fmt.Errorf("parse RBA investor housing credit: %w", err)
	}
	points := sumSeries(owner, investor)
	if len(points) == 0 {
		return nil, fmt.Errorf("no RBA housing credit observations")
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
		return nil, fmt.Errorf("fetch RBA: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch RBA: status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read RBA response: %w", err)
	}
	return body, nil
}

// sumSeries adds two series that share a timeline, keeping the dates they have
// in common.
func sumSeries(a, b []Point) []Point {
	byDate := make(map[time.Time]float64, len(b))
	for _, p := range b {
		byDate[p.Date] = p.Value
	}
	out := make([]Point, 0, len(a))
	for _, p := range a {
		v, ok := byDate[p.Date]
		if !ok {
			continue
		}
		out = append(out, Point{Date: p.Date, Value: p.Value + v})
	}
	return out
}

// parseColumn reads the column named title from an RBA statistics CSV body.
func parseColumn(body []byte, title string) ([]Point, error) {
	lines := strings.Split(string(body), "\n")

	headerAt, dataAt := -1, -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if headerAt < 0 && strings.HasPrefix(trimmed, "Title,") {
			headerAt = i
		}
		if strings.HasPrefix(trimmed, "Series ID,") {
			dataAt = i + 1
			break
		}
	}
	if headerAt < 0 || dataAt < 0 || dataAt >= len(lines) {
		return nil, fmt.Errorf("could not locate %q in RBA table", title)
	}

	titles, err := csv.NewReader(strings.NewReader(lines[headerAt])).Read()
	if err != nil {
		return nil, err
	}
	col := -1
	for i, t := range titles {
		if strings.TrimSpace(t) == title {
			col = i
			break
		}
	}
	if col < 0 {
		return nil, fmt.Errorf("column %q not found in header", title)
	}

	r := csv.NewReader(strings.NewReader(strings.Join(lines[dataAt:], "\n")))
	r.FieldsPerRecord = -1
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	var points []Point
	for _, row := range rows {
		if col >= len(row) {
			continue
		}
		t, err := time.Parse("02/01/2006", strings.TrimSpace(row[0]))
		if err != nil {
			continue
		}
		raw := strings.TrimSpace(row[col])
		if raw == "" {
			continue
		}
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			continue
		}
		points = append(points, Point{Date: t, Value: v})
	}
	if len(points) == 0 {
		return nil, fmt.Errorf("no observations for %q", title)
	}
	return points, nil
}
