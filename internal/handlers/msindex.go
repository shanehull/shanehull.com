package handlers

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/shanehull/shanehull.com/internal/cache"
	"github.com/shanehull/shanehull.com/internal/charts"
	"github.com/shanehull/shanehull.com/internal/fred"
	"github.com/shanehull/shanehull.com/internal/templates"
)

const (
	equityID = "NCBCEL"
	networthID = "TNWMVBSNNCB"
	cacheTTL = 24 * time.Hour
)

var chartCache = cache.New()

// getOrFetchChartData returns cached chart data or fetches and caches it
func getOrFetchChartData(rangeParam string, showQuartiles bool) ([]templates.LineChartData, error) {
	cacheKey := fmt.Sprintf("msindex:%s:%v", rangeParam, showQuartiles)

	// Check cache
	if cached, found := chartCache.Get(cacheKey); found {
		return cached.([]templates.LineChartData), nil
	}

	// Fetch data
	opts := &fred.FetchOptions{
		ObservationStart: charts.CalculateRangeStart(rangeParam),
		Frequency:        "q",
		Units:            "lin",
	}

	equityData, err := fred.FetchSeries(equityID, opts)
	if err != nil {
		return nil, err
	}

	networthData, err := fred.FetchSeries(networthID, opts)
	if err != nil {
		return nil, err
	}

	data := mergeAndCalculate(equityData, networthData)
	if len(data) == 0 {
		return nil, fmt.Errorf("no data available for the selected time range")
	}

	// Filter data by range if needed
	startDate := charts.CalculateRangeStart(rangeParam)
	if startDate != nil {
		filtered := make([]FinancialData, 0)
		for _, d := range data {
			if d.Date.After(*startDate) || d.Date.Equal(*startDate) {
				filtered = append(filtered, d)
			}
		}
		data = filtered
	}

	var q1, q3 []float64
	if showQuartiles {
		values := make([]float64, len(data))
		for i, d := range data {
			values[i] = d.MSIndex
		}
		q1, q3 = calculateQuartiles(values)
	}

	chartData := make([]templates.LineChartData, len(data))
	for i, d := range data {
		q1Val := 0.0
		q3Val := 0.0
		if showQuartiles && i < len(q1) && i < len(q3) {
			q1Val = q1[i]
			q3Val = q3[i]
		}
		chartData[i] = templates.LineChartData{
			Date:      d.Date.Format("2006-01-02"),
			Value:     d.MSIndex,
			Quartile1: q1Val,
			Quartile3: q3Val,
		}
	}

	// Cache the result
	chartCache.Set(cacheKey, chartData, cacheTTL)

	return chartData, nil
}

type FinancialData struct {
	Date     time.Time
	Equity   float64
	NetWorth float64
	MSIndex  float64
}



func MSIndexDownloadsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam := r.URL.Query().Get("range")
	if rangeParam == "" {
		rangeParam = "max"
	}

	showQuartiles := r.URL.Query().Has("quartiles")

	component := templates.ChartDownloads("msindex", rangeParam, "quartiles", showQuartiles)

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

func MSIndexCSVHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam := r.URL.Query().Get("range")
	if rangeParam == "" {
		rangeParam = "max"
	}

	showQuartiles := r.URL.Query().Has("quartiles")

	chartData, err := getOrFetchChartData(rangeParam, showQuartiles)
	if err != nil {
		log.Print("failed to get chart data:", err)
		http.Error(w, "Unable to load chart data. Please try again later.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=\"msindex-data.csv\"")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	header := []string{"date", "msindex"}
	if showQuartiles {
		header = append(header, "quartile1", "quartile3")
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

		if showQuartiles {
			row = append(row, fmt.Sprintf("%.6f", d.Quartile1), fmt.Sprintf("%.6f", d.Quartile3))
		}

		if err := writer.Write(row); err != nil {
			log.Print("failed to write CSV row:", err)
			return
		}
	}
}

func MSIndexDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam := r.URL.Query().Get("range")
	if rangeParam == "" {
		rangeParam = "max"
	}

	showQuartiles := r.URL.Query().Has("quartiles")

	chartData, err := getOrFetchChartData(rangeParam, showQuartiles)
	if err != nil {
		log.Print("failed to get chart data:", err)
		http.Error(w, "Unable to load chart data. Please try again later.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=\"msindex-data.json\"")

	if err := json.NewEncoder(w).Encode(chartData); err != nil {
		log.Print("failed to encode JSON:", err)
	}
}

func MSIndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam := r.URL.Query().Get("range")
	if rangeParam == "" {
		rangeParam = "max"
	}

	showQuartiles := r.URL.Query().Has("quartiles")

	chartData, err := getOrFetchChartData(rangeParam, showQuartiles)
	if err != nil {
		log.Print("failed to get chart data:", err)
		renderError(w, "Unable to load chart data. Please try again later.")
		return
	}

	options := map[string]string{
		"mainLabel":    "Misesian Stationarity Index",
		"yAxisLabel":   "Index Value",
		"showQuartiles": "false",
		"showAverage":  "false",
	}
	if showQuartiles {
		options["showQuartiles"] = "true"
	}
	component := templates.LineChart("chart-canvas", chartData, showQuartiles, options)

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



func mergeAndCalculate(equity, networth []fred.DataPoint) []FinancialData {
	networthMap := make(map[string]float64)
	for _, nw := range networth {
		networthMap[nw.Date.Format("2006-01-02")] = nw.Value
	}

	merged := make([]FinancialData, 0)
	for _, eq := range equity {
		dateStr := eq.Date.Format("2006-01-02")
		if nwValue, ok := networthMap[dateStr]; ok {
			merged = append(merged, FinancialData{
				Date:     eq.Date,
				Equity:   eq.Value,
				NetWorth: nwValue,
			})
		}
	}

	geometricScale(merged)
	return merged
}

func geometricScale(data []FinancialData) {
	if len(data) == 0 {
		return
	}

	product := 1.0
	for i, v := range data {
		unscaled := v.Equity / v.NetWorth
		product *= unscaled

		exponent := 1.0 / float64(i+1)
		geoMean := math.Pow(product, exponent)
		data[i].MSIndex = unscaled / geoMean
	}
}


