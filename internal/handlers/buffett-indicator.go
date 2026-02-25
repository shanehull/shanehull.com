package handlers

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/shanehull/shanehull.com/internal/cache"
	"github.com/shanehull/shanehull.com/internal/charts"
	"github.com/shanehull/shanehull.com/internal/fred"
	"github.com/shanehull/shanehull.com/internal/templates"
)

const (
	marketCapID    = "NCBEILQ027S"
	gdpID          = "GDP"
	buffetCacheTTL = 24 * time.Hour
)

var buffetCache = cache.New()

type BuffetData struct {
	Date      time.Time
	MarketCap float64
	GDP       float64
	Ratio     float64
}

func getOrFetchBuffetData(rangeParam string, showQuartiles bool) ([]templates.LineChartData, error) {
	cacheKey := fmt.Sprintf("buffet-indicator:%s", rangeParam)

	// Check cache
	if cached, found := buffetCache.Get(cacheKey); found {
		return cached.([]templates.LineChartData), nil
	}

	// Fetch data
	opts := &fred.FetchOptions{
		ObservationStart: charts.CalculateRangeStart(rangeParam),
		Frequency:        "q",
		Units:            "lin",
	}

	marketCapData, err := fred.FetchSeries(marketCapID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", marketCapID, err)
	}

	gdpData, err := fred.FetchSeries(gdpID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", gdpID, err)
	}

	data := mergeAndCalculateBuffet(marketCapData, gdpData)
	if len(data) == 0 {
		return nil, fmt.Errorf("no data available for the selected time range")
	}

	// Filter data by range if needed
	startDate := charts.CalculateRangeStart(rangeParam)
	if startDate != nil {
		filtered := make([]BuffetData, 0)
		for _, d := range data {
			if d.Date.After(*startDate) || d.Date.Equal(*startDate) {
				filtered = append(filtered, d)
			}
		}
		data = filtered
	}

	// Calculate average
	var avgRatio float64
	for _, d := range data {
		avgRatio += d.Ratio
	}
	avgRatio /= float64(len(data))

	chartData := make([]templates.LineChartData, len(data))
	for i, d := range data {
		chartData[i] = templates.LineChartData{
			Date:      d.Date.Format("2006-01-02"),
			Value:     d.Ratio,
			Quartile1: 0,
			Quartile3: 0,
			Average:   avgRatio,
		}
	}

	// Cache the result
	buffetCache.Set(cacheKey, chartData, buffetCacheTTL)

	return chartData, nil
}

func mergeAndCalculateBuffet(marketCap, gdp []fred.DataPoint) []BuffetData {
	gdpMap := make(map[string]float64)
	for _, g := range gdp {
		gdpMap[g.Date.Format("2006-01-02")] = g.Value
	}

	merged := make([]BuffetData, 0)
	for _, mc := range marketCap {
		dateStr := mc.Date.Format("2006-01-02")
		if gdpValue, ok := gdpMap[dateStr]; ok && gdpValue > 0 {
			ratio := (mc.Value / gdpValue) * 100 // as percentage
			merged = append(merged, BuffetData{
				Date:      mc.Date,
				MarketCap: mc.Value,
				GDP:       gdpValue,
				Ratio:     ratio,
			})
		}
	}

	return merged
}

func BuffettIndicatorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam := r.URL.Query().Get("range")
	if rangeParam == "" {
		rangeParam = "max"
	}

	showQuartiles := r.URL.Query().Has("quartiles")

	chartData, err := getOrFetchBuffetData(rangeParam, showQuartiles)
	if err != nil {
		log.Print("failed to get chart data:", err)
		renderError(w, "Unable to load chart data. Please try again later.")
		return
	}

	options := map[string]string{
		"mainLabel":     "Buffett Indicator",
		"yAxisLabel":    "Ratio (%)",
		"showAverage":   "true",
		"showQuartiles": "false",
	}
	component := templates.LineChart("chart-canvas", chartData, false, options)

	buf := new(bytes.Buffer)
	defer buf.Reset()

	if renderErr := component.Render(r.Context(), buf); renderErr != nil {
		log.Print("failed to render component:", renderErr)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("HX-Trigger", "initChartFromData")
	_, err = w.Write(buf.Bytes())
	if err != nil {
		log.Print("failed to write response:", err)
	}
}

func BuffettIndicatorDownloadsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam := r.URL.Query().Get("range")
	if rangeParam == "" {
		rangeParam = "max"
	}

	showQuartiles := r.URL.Query().Has("quartiles")

	component := templates.ChartDownloads("buffett-indicator", rangeParam, showQuartiles)

	buf := new(bytes.Buffer)
	defer buf.Reset()

	if err := component.Render(r.Context(), buf); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write(buf.Bytes()); err != nil {
		log.Print("failed to write response:", err)
	}
}

func BuffettIndicatorCSVHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam := r.URL.Query().Get("range")
	if rangeParam == "" {
		rangeParam = "max"
	}

	showQuartiles := r.URL.Query().Has("quartiles")

	chartData, err := getOrFetchBuffetData(rangeParam, showQuartiles)
	if err != nil {
		log.Print("failed to get chart data:", err)
		http.Error(w, "Unable to load chart data. Please try again later.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=\"buffet-indicator-data.csv\"")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	header := []string{"date", "buffett_indicator", "average"}
	if err := writer.Write(header); err != nil {
		log.Print("failed to write CSV header:", err)
		return
	}

	for _, d := range chartData {
		row := []string{
			d.Date,
			fmt.Sprintf("%.6f", d.Value),
			fmt.Sprintf("%.6f", d.Average),
		}

		if err := writer.Write(row); err != nil {
			log.Print("failed to write CSV row:", err)
			return
		}
	}
}

func BuffettIndicatorDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam := r.URL.Query().Get("range")
	if rangeParam == "" {
		rangeParam = "max"
	}

	showQuartiles := r.URL.Query().Has("quartiles")

	chartData, err := getOrFetchBuffetData(rangeParam, showQuartiles)
	if err != nil {
		log.Print("failed to get chart data:", err)
		http.Error(w, "Unable to load chart data. Please try again later.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=\"buffet-indicator-data.json\"")

	if err := json.NewEncoder(w).Encode(chartData); err != nil {
		log.Print("failed to encode JSON:", err)
	}
}
