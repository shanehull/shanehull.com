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
	federalDebtID  = "GFDEBTN"
	buffetCacheTTL = 24 * time.Hour
)

var buffetCache = cache.New()

type BuffetData struct {
	Date      time.Time
	MarketCap float64
	GDP       float64
	Debt      float64
	Ratio     float64
}

func getOrFetchBuffetData(rangeParam string, showAverage, includeDebt bool) ([]templates.LineChartData, error) {
	cacheKey := fmt.Sprintf("buffet-indicator:%s:%v:%v", rangeParam, showAverage, includeDebt)

	// Check cache
	if cached, found := buffetCache.Get(cacheKey); found {
		return cached.([]templates.LineChartData), nil
	}

	// Fetch the series concurrently; FRED latency dominates the cold load, so
	// parallel round trips keep it to one round trip's wait.
	series := []string{marketCapID, gdpID, federalDebtID}
	results, errs := fetchAll(len(series), func(i int) ([]fred.DataPoint, error) {
		opts := &fred.FetchOptions{
			ObservationStart: charts.CalculateRangeStart(rangeParam),
			Frequency:        "q",
			Units:            "lin",
		}
		data, err := fred.FetchSeries(series[i], opts)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch %s: %w", series[i], err)
		}
		return data, nil
	})
	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}

	data := mergeAndCalculateBuffet(results[0], results[1], results[2], includeDebt)
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
		avg := 0.0
		if showAverage {
			avg = avgRatio
		}
		chartData[i] = templates.LineChartData{
			Date:      d.Date.Format("2006-01-02"),
			Value:     d.Ratio,
			Quartile1: 0,
			Quartile3: 0,
			Average:   avg,
		}
	}

	// Cache the result
	buffetCache.Set(cacheKey, chartData, buffetCacheTTL)

	return chartData, nil
}

func mergeAndCalculateBuffet(
	marketCap, gdp, debt []fred.DataPoint,
	includeDebt bool,
) []BuffetData {
	gdpMap := make(map[string]float64)
	for _, g := range gdp {
		gdpMap[g.Date.Format("2006-01-02")] = g.Value
	}

	debtMap := make(map[string]float64)
	for _, d := range debt {
		debtMap[d.Date.Format("2006-01-02")] = d.Value
	}

	merged := make([]BuffetData, 0)
	for _, mc := range marketCap {
		dateStr := mc.Date.Format("2006-01-02")
		gdpValue, ok := gdpMap[dateStr]
		if !ok || gdpValue <= 0 {
			continue
		}

		debtValue, hasDebt := debtMap[dateStr]
		if !hasDebt {
			continue
		}

		numerator := mc.Value / 1000
		if includeDebt {
			numerator += debtValue / 1000
		}

		merged = append(merged, BuffetData{
			Date:      mc.Date,
			MarketCap: mc.Value,
			GDP:       gdpValue,
			Debt:      debtValue,
			Ratio:     (numerator / gdpValue) * 100,
		})
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

	showAverage := r.URL.Query().Get("average") == "on"
	showDebt := r.URL.Query().Get("debt") == "on"

	chartData, err := getOrFetchBuffetData(rangeParam, showAverage, showDebt)
	if err != nil {
		log.Print("failed to get chart data:", err)
		renderError(w, "Unable to load chart data. Please try again later.")
		return
	}

	mainLabel := "Buffett Indicator"
	if showDebt {
		mainLabel = "Buffett Indicator (incl. federal debt)"
	}

	options := map[string]string{
		"mainLabel":     mainLabel,
		"yAxisLabel":    "Ratio (%)",
		"showQuartiles": "false",
	}
	if showAverage {
		options["showAverage"] = "true"
	} else {
		options["showAverage"] = "false"
	}
	component := templates.LineChart("chart-canvas", chartData, false, options)

	buf := new(bytes.Buffer)
	defer buf.Reset()

	if renderErr := component.Render(r.Context(), buf); renderErr != nil {
		// Suppress context canceled errors - common when client disconnects
		if renderErr.Error() != "context canceled" {
			log.Print("failed to render component:", renderErr)
		}
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

	showAverage := r.URL.Query().Get("average") == "on"
	showDebt := r.URL.Query().Get("debt") == "on"

	component := templates.ChartDownloads("buffett-indicator", rangeParam, map[string]string{
		"average": templates.Param(showAverage),
		"debt":    templates.Param(showDebt),
	})

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

	showAverage := r.URL.Query().Get("average") == "on"
	showDebt := r.URL.Query().Get("debt") == "on"

	chartData, err := getOrFetchBuffetData(rangeParam, showAverage, showDebt)
	if err != nil {
		log.Print("failed to get chart data:", err)
		http.Error(w, "Unable to load chart data. Please try again later.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=\"buffet-indicator-data.csv\"")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	header := []string{"date", "buffett_indicator"}
	if showAverage {
		header = append(header, "average")
	}
	if err := writer.Write(header); err != nil {
		log.Print("failed to write CSV header:", err)
		return
	}

	for _, d := range chartData {
		row := []string{
			d.Date,
			fmt.Sprintf("%.6f", d.Value),
		}

		if showAverage {
			row = append(row, fmt.Sprintf("%.6f", d.Average))
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

	showAverage := r.URL.Query().Get("average") == "on"
	showDebt := r.URL.Query().Get("debt") == "on"

	chartData, err := getOrFetchBuffetData(rangeParam, showAverage, showDebt)
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
