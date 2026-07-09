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
	tbillID          = "TB3MS"
	cpiID            = "CPIAUCNS"
	realRateCacheTTL = 24 * time.Hour
)

var realRateCache = cache.New()

func getOrFetchRealRateData(rangeParam string, showAverage bool) ([]templates.LineChartData, error) {
	cacheKey := fmt.Sprintf("real-interest-rate:%s:%v", rangeParam, showAverage)

	if cached, found := realRateCache.Get(cacheKey); found {
		return cached.([]templates.LineChartData), nil
	}

	rangeStart := charts.CalculateRangeStart(rangeParam)

	cpiStart := rangeStart
	if cpiStart != nil {
		earlier := cpiStart.AddDate(-1, 0, 0)
		cpiStart = &earlier
	}

	opts := &fred.FetchOptions{
		ObservationStart: rangeStart,
		Frequency:        "m",
		Units:            "lin",
	}
	cpiOpts := &fred.FetchOptions{
		ObservationStart: cpiStart,
		Frequency:        "m",
		Units:            "lin",
	}

	tbillData, err := fred.FetchSeries(tbillID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", tbillID, err)
	}

	cpiData, err := fred.FetchSeries(cpiID, cpiOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", cpiID, err)
	}

	cpiMap := make(map[string]float64)
	for _, c := range cpiData {
		cpiMap[c.Date.Format("2006-01-02")] = c.Value
	}

	type realRatePoint struct {
		Date     time.Time
		TBill    float64
		CPIYoY   float64
		RealRate float64
	}

	points := make([]realRatePoint, 0)
	for _, tb := range tbillData {
		cpiNow, ok := cpiMap[tb.Date.Format("2006-01-02")]
		if !ok {
			continue
		}
		yearAgo := tb.Date.AddDate(-1, 0, 0)
		cpiYearAgo, ok := cpiMap[yearAgo.Format("2006-01-02")]
		if !ok {
			continue
		}
		if cpiYearAgo == 0 {
			continue
		}

		cpiYoY := ((cpiNow / cpiYearAgo) - 1) * 100
		realRate := tb.Value - cpiYoY

		points = append(points, realRatePoint{
			Date:     tb.Date,
			TBill:    tb.Value,
			CPIYoY:   cpiYoY,
			RealRate: realRate,
		})
	}

	if rangeStart != nil {
		filtered := make([]realRatePoint, 0)
		for _, p := range points {
			if p.Date.After(*rangeStart) || p.Date.Equal(*rangeStart) {
				filtered = append(filtered, p)
			}
		}
		points = filtered
	}

	if len(points) == 0 {
		return nil, fmt.Errorf("no data available for the selected time range")
	}

	var avgRate float64
	for _, p := range points {
		avgRate += p.RealRate
	}
	avgRate /= float64(len(points))

	chartData := make([]templates.LineChartData, len(points))
	for i, p := range points {
		avg := 0.0
		if showAverage {
			avg = avgRate
		}
		chartData[i] = templates.LineChartData{
			Date:    p.Date.Format("2006-01-02"),
			Value:   p.RealRate,
			Average: avg,
		}
	}

	realRateCache.Set(cacheKey, chartData, realRateCacheTTL)

	return chartData, nil
}

func RealInterestRateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam := r.URL.Query().Get("range")
	if rangeParam == "" {
		rangeParam = "max"
	}

	showAverage := r.URL.Query().Has("average")

	chartData, err := getOrFetchRealRateData(rangeParam, showAverage)
	if err != nil {
		log.Print("failed to get chart data:", err)
		renderError(w, "Unable to load chart data. Please try again later.")
		return
	}

	options := map[string]string{
		"mainLabel":     "Real T-Bill Rate (3-Mo T-Bill - CPI YoY%)",
		"yAxisLabel":    "Rate (%)",
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

func RealInterestRateDownloadsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam := r.URL.Query().Get("range")
	if rangeParam == "" {
		rangeParam = "max"
	}

	showAverage := r.URL.Query().Has("average")

	component := templates.ChartDownloads("real-interest-rate", rangeParam, "average", showAverage)

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

func RealInterestRateCSVHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam := r.URL.Query().Get("range")
	if rangeParam == "" {
		rangeParam = "max"
	}

	showAverage := r.URL.Query().Has("average")

	chartData, err := getOrFetchRealRateData(rangeParam, showAverage)
	if err != nil {
		log.Print("failed to get chart data:", err)
		http.Error(w, "Unable to load chart data. Please try again later.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=\"real-interest-rate-data.csv\"")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	header := []string{"date", "real_interest_rate"}
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

func RealInterestRateDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam := r.URL.Query().Get("range")
	if rangeParam == "" {
		rangeParam = "max"
	}

	showAverage := r.URL.Query().Has("average")

	chartData, err := getOrFetchRealRateData(rangeParam, showAverage)
	if err != nil {
		log.Print("failed to get chart data:", err)
		http.Error(w, "Unable to load chart data. Please try again later.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=\"real-interest-rate-data.json\"")

	if err := json.NewEncoder(w).Encode(chartData); err != nil {
		log.Print("failed to encode JSON:", err)
	}
}
