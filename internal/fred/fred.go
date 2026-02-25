package fred

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

const baseURL = "https://api.stlouisfed.org/fred/series/observations"

// Observation represents a single data point from FRED
type Observation struct {
	Date  string `json:"date"`
	Value string `json:"value"`
}

// Response represents the FRED API response
type Response struct {
	Observations []Observation `json:"observations"`
}

// DataPoint represents a parsed observation with a time.Time
type DataPoint struct {
	Date  time.Time
	Value float64
}

// FetchOptions configures parameters for fetching FRED series data.
// All fields are optional; reasonable defaults are provided.
type FetchOptions struct {
	// ObservationStart limits results to on or after this date (YYYY-MM-DD).
	// If nil, FRED returns all available data for the series.
	ObservationStart *time.Time

	// ObservationEnd limits results to on or before this date (YYYY-MM-DD).
	// Defaults to today if not set.
	ObservationEnd *time.Time

	// Frequency aggregates data: "d" (daily), "w" (weekly), "bw" (biweekly),
	// "m" (monthly), "q" (quarterly), "sa" (semi-annual), "a" (annual).
	// Defaults to "q" (quarterly) if not set.
	Frequency string

	// Units specifies the transformation: "lin" (linear), "chg" (change),
	// "ch1" (change from 1 year ago), "pch" (percent change),
	// "pc1" (percent change from 1 year ago), "pca" (percent change annual rate).
	// Defaults to "lin" (linear) if not set.
	Units string

	// Limit the maximum number of observations returned (1-100000).
	// If not set or 0, no limit is applied.
	Limit int

	// Offset the starting row of results (0-indexed).
	// Defaults to 0 if not set.
	Offset int

	// SortOrder determines result order: "asc" (ascending) or "desc" (descending).
	// Defaults to "asc" if not set.
	SortOrder string
}

// applyDefaults fills in missing FetchOptions with sensible defaults
func (opts *FetchOptions) applyDefaults() {
	if opts.ObservationEnd == nil {
		end := time.Now()
		opts.ObservationEnd = &end
	}
	if opts.Frequency == "" {
		opts.Frequency = "q"
	}
	if opts.Units == "" {
		opts.Units = "lin"
	}
	if opts.SortOrder == "" {
		opts.SortOrder = "asc"
	}
}

// FetchSeries retrieves observations for a given FRED series ID.
// The opts parameter can be nil to use sensible defaults.
func FetchSeries(seriesID string, opts *FetchOptions) ([]DataPoint, error) {
	apiKey := os.Getenv("FRED_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("FRED_API_KEY environment variable not set")
	}

	if opts == nil {
		opts = &FetchOptions{}
	}
	opts.applyDefaults()

	// Build query parameters
	query := url.Values{}
	query.Set("series_id", seriesID)
	query.Set("api_key", apiKey)
	query.Set("file_type", "json")
	if opts.ObservationStart != nil {
		query.Set("observation_start", opts.ObservationStart.Format("2006-01-02"))
	}
	query.Set("observation_end", opts.ObservationEnd.Format("2006-01-02"))
	query.Set("frequency", opts.Frequency)
	query.Set("units", opts.Units)
	query.Set("sort_order", opts.SortOrder)

	if opts.Limit > 0 {
		query.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("offset", strconv.Itoa(opts.Offset))
	}

	requestURL := baseURL + "?" + query.Encode()

	resp, err := http.Get(requestURL)
	if err != nil {
		// Provide better error messages for common network issues
		if _, ok := err.(net.Error); ok {
			return nil, fmt.Errorf("network error fetching FRED series %s: %w", seriesID, err)
		}
		return nil, fmt.Errorf("failed to fetch FRED series %s: %w", seriesID, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("FRED API error (status %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read FRED response: %w", err)
	}

	var fredResp Response
	if err := json.Unmarshal(body, &fredResp); err != nil {
		return nil, fmt.Errorf("failed to parse FRED response: %w", err)
	}

	data := make([]DataPoint, 0, len(fredResp.Observations))
	for _, obs := range fredResp.Observations {
		// Skip missing values represented as "."
		if obs.Value == "." {
			continue
		}

		parsedDate, err := time.Parse("2006-01-02", obs.Date)
		if err != nil {
			continue
		}

		value, err := strconv.ParseFloat(obs.Value, 64)
		if err != nil {
			continue
		}

		data = append(data, DataPoint{
			Date:  parsedDate,
			Value: value,
		})
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("no data available for series %s in the requested time range", seriesID)
	}

	return data, nil
}
