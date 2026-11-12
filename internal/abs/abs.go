// Package abs fetches Australian Bureau of Statistics series. The client is
// stateless; callers apply their own caching.
package abs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
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
	defaultAPIURL = "https://data.api.abs.gov.au"
	earningsKey   = "/rest/data/ABS,AWE,1.0.0/1.1.3.7.TOT.10.AUS.S?format=jsondata"
)

// Client fetches ABS data.
type Client struct {
	apiURL     string
	httpClient *http.Client
}

// Option configures a Client.
type Option func(*Client)

// WithAPIURL overrides the SDMX data API endpoint. Used by tests.
func WithAPIURL(rawURL string) Option {
	return func(c *Client) { c.apiURL = rawURL }
}

// WithHTTPClient overrides the HTTP client, including its timeout.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) { c.httpClient = client }
}

// New returns a Client with sensible defaults.
func New(opts ...Option) *Client {
	c := &Client{
		apiURL:     defaultAPIURL,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// FetchAverageWeeklyEarnings returns average weekly earnings for all employees,
// persons, in dollars, half-yearly and oldest first.
func (c *Client) FetchAverageWeeklyEarnings(ctx context.Context) ([]Point, error) {
	body, err := c.get(ctx, c.apiURL+earningsKey, "application/vnd.sdmx.data+json")
	if err != nil {
		return nil, err
	}
	observations, err := parseSDMX(body)
	if err != nil {
		return nil, fmt.Errorf("parse ABS average weekly earnings: %w", err)
	}
	points, err := toPoints(observations)
	if err != nil {
		return nil, fmt.Errorf("parse ABS average weekly earnings: %w", err)
	}
	if len(points) == 0 {
		return nil, fmt.Errorf("no ABS average weekly earnings observations")
	}
	return points, nil
}

func (c *Client) get(ctx context.Context, endpoint, accept string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
	if accept != "" {
		req.Header.Set("Accept", accept)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch ABS: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch ABS: status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read ABS response: %w", err)
	}
	return body, nil
}

// sdmxDimension is a dimension with its ordered value ids.
type sdmxDimension struct {
	Values []struct {
		ID string `json:"id"`
	} `json:"values"`
}

type sdmxResponse struct {
	Data struct {
		Structures []struct {
			Dimensions struct {
				Observation []sdmxDimension `json:"observation"`
			} `json:"dimensions"`
		} `json:"structures"`
		DataSets []struct {
			Series map[string]struct {
				Observations map[string][]json.RawMessage `json:"observations"`
			} `json:"series"`
		} `json:"dataSets"`
	} `json:"data"`
	Errors []struct {
		Title string `json:"title"`
	} `json:"errors"`
}

// parseSDMX reads ABS SDMX-JSON observations keyed by period. Each observation
// value is the first element of its tuple.
func parseSDMX(raw []byte) (map[string]float64, error) {
	var resp sdmxResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("decode SDMX: %w", err)
	}
	if len(resp.Errors) > 0 {
		return nil, fmt.Errorf("SDMX error: %s", resp.Errors[0].Title)
	}
	if len(resp.Data.Structures) == 0 || len(resp.Data.DataSets) == 0 {
		return nil, fmt.Errorf("empty SDMX response")
	}

	var periods []string
	if dims := resp.Data.Structures[0].Dimensions.Observation; len(dims) > 0 {
		for _, v := range dims[0].Values {
			periods = append(periods, v.ID)
		}
	}

	observations := map[string]float64{}
	for _, series := range resp.Data.DataSets[0].Series {
		for key, tuple := range series.Observations {
			i, err := strconv.Atoi(key)
			if err != nil || i >= len(periods) || len(tuple) == 0 {
				continue
			}
			v, err := parseNumber(tuple[0])
			if err != nil {
				continue
			}
			observations[periods[i]] = v
		}
	}
	return observations, nil
}

func parseNumber(raw json.RawMessage) (float64, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return 0, fmt.Errorf("null")
	}
	var v float64
	if err := json.Unmarshal(raw, &v); err != nil {
		return 0, err
	}
	return v, nil
}

func toPoints(observations map[string]float64) ([]Point, error) {
	points := make([]Point, 0, len(observations))
	for period, value := range observations {
		t, err := parsePeriod(period)
		if err != nil {
			continue
		}
		points = append(points, Point{Date: t, Value: value})
	}
	sort.Slice(points, func(i, j int) bool { return points[i].Date.Before(points[j].Date) })
	return points, nil
}

// parsePeriod converts an ABS period string into a period-start date.
//
//	"2026-Q2" -> 2026-04-01
//	"2026-S1" -> 2026-05-01 (half-yearly, dated at its midpoint)
//	"2026-S2" -> 2026-11-01
func parsePeriod(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("empty period")
	}
	if strings.ContainsAny(s, "Qq") {
		year, index, err := parseNumbered(s, "Q")
		if err != nil {
			return time.Time{}, err
		}
		return time.Date(year, time.Month((index-1)*3+1), 1, 0, 0, 0, 0, time.UTC), nil
	}
	if strings.ContainsAny(s, "Ss") {
		year, half, err := parseNumbered(s, "S")
		if err != nil {
			return time.Time{}, err
		}
		month := 5
		if half == 2 {
			month = 11
		}
		return time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC), nil
	}
	return time.Time{}, fmt.Errorf("unrecognised period %q", s)
}

func parseNumbered(s, marker string) (year, index int, err error) {
	i := strings.IndexAny(s, strings.ToUpper(marker)+strings.ToLower(marker))
	if i < 0 {
		return 0, 0, fmt.Errorf("not %s: %q", marker, s)
	}
	year, err = strconv.Atoi(strings.TrimRight(strings.TrimSpace(s[:i]), "-_ "))
	if err != nil {
		return 0, 0, err
	}
	index, err = strconv.Atoi(strings.TrimSpace(s[i+1:]))
	if err != nil {
		return 0, 0, err
	}
	return year, index, nil
}
