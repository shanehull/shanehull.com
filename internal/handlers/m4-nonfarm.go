package handlers

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/shanehull/shanehull.com/internal/cache"
	"github.com/shanehull/shanehull.com/internal/cfs"
	"github.com/shanehull/shanehull.com/internal/charts"
	"github.com/shanehull/shanehull.com/internal/fred"
	"github.com/shanehull/shanehull.com/internal/templates"
)

// Money vs nonfarm productivity growth: the quantity-theory gauge of how much
// broad money is being created relative to real output per hour. Divisia M4
// (the broadest honest money) growth minus nonfarm labor productivity growth
// is the inflation pressure that real assets, gold first among them, catch.
// When money grows faster than productivity the excess line is positive.

const (
	productivityID = "OPHNFB" // Nonfarm output per hour, index 2017=100
	m4NonfarmTTL   = 24 * time.Hour
)

var m4NonfarmCache = cache.New()

type m4NonfarmPoint struct {
	Date    time.Time
	Money   float64 // Divisia M4 year-over-year growth (%)
	Product float64 // Nonfarm productivity year-over-year growth (%)
	Excess  float64 // Money - Product
}

// buildM4Nonfarm aligns quarterly productivity observations with the Divisia
// M4 index and computes both on the same year-over-year basis: each series
// against its level four quarters earlier. Averaging quarterly annualized
// rates would not equal compounding, so year-over-year is the honest yearly
// measure and is smooth enough that a single covid quarter cannot flatten the
// chart.
func buildM4Nonfarm(m4 []cfs.Point, prod []fred.DataPoint) []m4NonfarmPoint {
	m4ByMonth := make(map[string]float64, len(m4))
	for _, p := range m4 {
		m4ByMonth[p.Date.Format("2006-01")] = p.Index
	}

	prodByDate := make(map[string]float64, len(prod))
	for _, p := range prod {
		prodByDate[p.Date.Format("2006-01-02")] = p.Value
	}

	points := make([]m4NonfarmPoint, 0, len(prod))
	for _, p := range prod {
		d := p.Date

		now, ok1 := m4ByMonth[d.Format("2006-01")]
		prevY, ok2 := m4ByMonth[d.AddDate(-1, 0, 0).Format("2006-01")]
		prodPrev, ok3 := prodByDate[d.AddDate(0, -12, 0).Format("2006-01-02")]
		if !ok1 || !ok2 || !ok3 || prevY <= 0 || prodPrev <= 0 {
			continue
		}

		money := (now/prevY - 1) * 100
		product := (p.Value/prodPrev - 1) * 100
		points = append(points, m4NonfarmPoint{
			Date:    d,
			Money:   money,
			Product: product,
			Excess:  money - product,
		})
	}

	sort.Slice(points, func(i, j int) bool { return points[i].Date.Before(points[j].Date) })
	return points
}

type m4NonfarmResult struct {
	Labels  []string
	Money   []*float64
	Product []*float64
	Excess  []*float64
}

func getM4Nonfarm(rangeParam string) (*m4NonfarmResult, error) {
	cacheKey := "m4-nonfarm:" + rangeParam
	if cached, found := m4NonfarmCache.Get(cacheKey); found {
		return cached.(*m4NonfarmResult), nil
	}

	m4, err := fetchM4Index()
	if err != nil {
		return nil, err
	}

	opts := &fred.FetchOptions{Frequency: "q", Units: "lin"}
	prod, err := fred.FetchSeries(productivityID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", productivityID, err)
	}

	points := buildM4Nonfarm(m4, prod)
	rangeStart := charts.CalculateRangeStart(rangeParam)

	result := &m4NonfarmResult{}
	for _, p := range points {
		if rangeStart != nil && p.Date.Before(*rangeStart) {
			continue
		}
		money := p.Money
		product := p.Product
		excess := p.Excess
		result.Labels = append(result.Labels, p.Date.Format("2006-01-02"))
		result.Money = append(result.Money, &money)
		result.Product = append(result.Product, &product)
		result.Excess = append(result.Excess, &excess)
	}

	if len(result.Labels) == 0 {
		return nil, fmt.Errorf("no data available for the selected time range")
	}

	m4NonfarmCache.Set(cacheKey, result, m4NonfarmTTL)
	return result, nil
}

func requestM4NFRange(r *http.Request) string {
	rangeParam := r.URL.Query().Get("range")
	if rangeParam == "" {
		rangeParam = "max"
	}
	return rangeParam
}

// M4NonfarmHandler renders the money vs productivity chart.
func M4NonfarmHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam := requestM4NFRange(r)
	result, err := getM4Nonfarm(rangeParam)
	if err != nil {
		log.Print("failed to get m4-nonfarm data:", err)
		renderError(w, "Unable to load chart data. Please try again later.")
		return
	}

	chartData := make([]templates.LineChartData, 0, len(result.Labels))
	for i, label := range result.Labels {
		chartData = append(chartData, templates.LineChartData{
			Date:  label,
			Value: *result.Excess[i],
		})
	}

	options := map[string]string{
		"mainLabel":     "Excess: M4 growth minus productivity (YoY)",
		"yAxisLabel":    "Percentage points (year over year)",
		"showQuartiles": "false",
		"showAverage":   "false",
	}
	component := templates.LineChart("chart-canvas", chartData, false, options)

	buf := new(bytes.Buffer)
	defer buf.Reset()

	if err := component.Render(r.Context(), buf); err != nil {
		log.Print("failed to render component:", err)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("HX-Trigger", "initChartFromData")
	if _, err = w.Write(buf.Bytes()); err != nil {
		log.Print("failed to write response:", err)
	}
}

// M4NonfarmDownloadsHandler renders the JSON/CSV download links.
func M4NonfarmDownloadsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam := requestM4NFRange(r)
	component := templates.MultiChartDownloads("m4-nonfarm", rangeParam, []string{"m4", "nonfarm", "excess"})

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

// M4NonfarmDataHandler returns the quarterly values as JSON.
func M4NonfarmDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam := requestM4NFRange(r)
	result, err := getM4Nonfarm(rangeParam)
	if err != nil {
		log.Print("failed to get m4-nonfarm data:", err)
		http.Error(w, "Unable to load chart data. Please try again later.", http.StatusInternalServerError)
		return
	}

	rows := make([]map[string]any, 0, len(result.Labels))
	for i, label := range result.Labels {
		row := map[string]any{
			"date":    label,
			"m4":      *result.Money[i],
			"nonfarm": *result.Product[i],
			"excess":  *result.Excess[i],
		}
		rows = append(rows, row)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"m4-nonfarm-data.json\"")
	if err := json.NewEncoder(w).Encode(rows); err != nil {
		log.Print("failed to encode JSON:", err)
	}
}

// M4NonfarmCSVHandler returns the quarterly values as CSV.
func M4NonfarmCSVHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam := requestM4NFRange(r)
	result, err := getM4Nonfarm(rangeParam)
	if err != nil {
		log.Print("failed to get m4-nonfarm data:", err)
		http.Error(w, "Unable to load chart data. Please try again later.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"m4-nonfarm-data.csv\"")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	if err := writer.Write([]string{"date", "m4", "nonfarm", "excess"}); err != nil {
		log.Print("failed to write CSV header:", err)
		return
	}
	for i, label := range result.Labels {
		row := []string{
			label,
			strings.TrimRight(fmt.Sprintf("%.2f", *result.Money[i]), "0"),
			strings.TrimRight(fmt.Sprintf("%.2f", *result.Product[i]), "0"),
			strings.TrimRight(fmt.Sprintf("%.2f", *result.Excess[i]), "0"),
		}
		if err := writer.Write(row); err != nil {
			log.Print("failed to write CSV row:", err)
			return
		}
	}
}
