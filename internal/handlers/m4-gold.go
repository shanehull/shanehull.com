package handlers

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/shanehull/shanehull.com/internal/cache"
	"github.com/shanehull/shanehull.com/internal/cfs"
	"github.com/shanehull/shanehull.com/internal/charts"
	"github.com/shanehull/shanehull.com/internal/data"
	"github.com/shanehull/shanehull.com/internal/templates"
	"github.com/shanehull/shanehull.com/internal/yahoo"
)

// Divisia M4 to gold, the broadest honest money aggregate measured per ounce
// of gold. The Center for Financial Stability publishes Divisia M4 as a
// quantity index (1967 = 100), not a dollar level, and FRED carries no gold
// price; so both legs are indexed to 1967 and the ratio is
//
//	M4-per-ounce(t) = M4-index(t) x gold-price(1967) / gold-price(t)
//
// which cancels the unknown dollar base level exactly. A rising line means
// broad money is growing faster per ounce than gold, i.e. monetary tinder
// building against the metal; a falling line means gold is revaluing faster
// than money. M2 misses the institutional layer and Fed M3 died in 2006,
// which is why Divisia M4 is the right numerator for tinder.

const (
	m4GoldTTL    = 24 * time.Hour
	goldPriceKey = "m4-gold:gold-price"
	m4IndexKey   = "m4-gold:m4-index"
)

var m4GoldCache = cache.New()
var goldPriceCache = cache.New()

type goldPricePoint struct {
	Date  time.Time
	Price float64
}

// loadGoldPriceSeries returns the monthly gold price from 1960 to today.
// History is embedded (World Bank Pink Sheet for 1960-2024); a live Yahoo
// GC=F tail extends past the last embedded month, scaled so levels stay
// continuous. If the live tail is unavailable the embedded history is used
// alone. Cached for 24 hours.
func loadGoldPriceSeries() ([]goldPricePoint, error) {
	if cached, found := goldPriceCache.Get(goldPriceKey); found {
		return cached.([]goldPricePoint), nil
	}

	points, err := parseGoldPriceCSV(data.GoldPriceCSV)
	if err != nil {
		return nil, err
	}
	last := points[len(points)-1]

	history := make(map[string]float64, len(points))
	for _, p := range points {
		history[p.Date.Format("2006-01")] = p.Price
	}

	live, err := yahoo.New().FetchMonthly("GC=F")
	if err != nil {
		log.Print("failed to fetch live gold price tail, using embedded history:", err)
		goldPriceCache.Set(goldPriceKey, points, m4GoldTTL)
		return points, nil
	}

	var sum, n float64
	for _, p := range live {
		if v, ok := history[p.Date.Format("2006-01")]; ok {
			sum += v / p.Price
			n++
		}
	}
	factor := sum / n

	merged := make([]goldPricePoint, 0, len(points)+len(live))
	merged = append(merged, points...)
	for _, p := range live {
		if p.Date.After(last.Date) {
			merged = append(merged, goldPricePoint{Date: p.Date, Price: p.Price * factor})
		}
	}

	goldPriceCache.Set(goldPriceKey, merged, m4GoldTTL)
	return merged, nil
}

// parseGoldPriceCSV turns the embedded monthly gold CSV into price points.
func parseGoldPriceCSV(csvBytes []byte) ([]goldPricePoint, error) {
	reader := csv.NewReader(bytes.NewReader(csvBytes))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parse gold price CSV: %w", err)
	}

	points := make([]goldPricePoint, 0, len(records)-1)
	for _, row := range records[1:] { // skip header
		date, err := time.Parse("2006-01", row[0])
		if err != nil {
			continue
		}
		price, err := strconv.ParseFloat(row[1], 64)
		if err != nil {
			continue
		}
		points = append(points, goldPricePoint{Date: date, Price: price})
	}

	if len(points) == 0 {
		return nil, fmt.Errorf("gold price CSV is empty")
	}
	return points, nil
}

type m4GoldPoint struct {
	Date  time.Time
	Index float64 // Divisia M4 per ounce of gold, 1967 = 100
}

// mergeM4Gold aligns the Divisia M4 index with the monthly gold price and
// expresses M4 per ounce of gold indexed to 1967. The 1967 gold price anchors
// the ratio, cancelling the unknown dollar base of the Divisia index.
func mergeM4Gold(m4 []cfs.Point, gold []goldPricePoint) []m4GoldPoint {
	goldByMonth := make(map[string]float64, len(gold))
	for _, g := range gold {
		goldByMonth[g.Date.Format("2006-01")] = g.Price
	}

	price1967, ok := goldByMonth["1967-01"]
	if !ok {
		// Fall back to the earliest available gold price so the ratio still
		// has a base even if the 1967 month is missing.
		for _, g := range gold {
			if g.Price > 0 {
				price1967 = g.Price
				break
			}
		}
	}

	points := make([]m4GoldPoint, 0, len(m4))
	for _, d := range m4 {
		price, ok := goldByMonth[d.Date.Format("2006-01")]
		if !ok || price <= 0 {
			continue
		}
		// index(t) x gold(1967) / gold(t); 1967 maps to 100.
		points = append(points, m4GoldPoint{Date: d.Date, Index: d.Index * price1967 / price})
	}

	sort.Slice(points, func(i, j int) bool { return points[i].Date.Before(points[j].Date) })
	return points
}

// fetchM4Index returns the monthly CFS Divisia M4 index, cached for 24 hours.
func fetchM4Index() ([]cfs.Point, error) {
	if cached, found := m4GoldCache.Get(m4IndexKey); found {
		return cached.([]cfs.Point), nil
	}

	points, err := cfs.New().FetchM4Index()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Divisia M4: %w", err)
	}

	m4GoldCache.Set(m4IndexKey, points, m4GoldTTL)
	return points, nil
}

func getM4Gold(rangeParam string) ([]templates.LineChartData, error) {
	cacheKey := "m4-gold:" + rangeParam
	if cached, found := m4GoldCache.Get(cacheKey); found {
		return cached.([]templates.LineChartData), nil
	}

	gold, err := loadGoldPriceSeries()
	if err != nil {
		return nil, fmt.Errorf("failed to load gold price: %w", err)
	}
	m4, err := fetchM4Index()
	if err != nil {
		return nil, err
	}

	points := mergeM4Gold(m4, gold)
	rangeStart := charts.CalculateRangeStart(rangeParam)

	chartData := make([]templates.LineChartData, 0, len(points))
	for _, p := range points {
		if rangeStart != nil && p.Date.Before(*rangeStart) {
			continue
		}
		chartData = append(chartData, templates.LineChartData{
			Date:  p.Date.Format("2006-01-02"),
			Value: p.Index,
		})
	}

	if len(chartData) == 0 {
		return nil, fmt.Errorf("no data available for the selected time range")
	}

	m4GoldCache.Set(cacheKey, chartData, m4GoldTTL)
	return chartData, nil
}

func requestM4Range(r *http.Request) string {
	rangeParam := r.URL.Query().Get("range")
	if rangeParam == "" {
		rangeParam = "max"
	}
	return rangeParam
}

// M4GoldHandler renders the Divisia M4-to-gold chart.
func M4GoldHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam := requestM4Range(r)
	chartData, err := getM4Gold(rangeParam)
	if err != nil {
		log.Print("failed to get m4-gold data:", err)
		renderError(w, "Unable to load chart data. Please try again later.")
		return
	}

	options := map[string]string{
		"mainLabel":     "Divisia M4 / Gold",
		"yAxisLabel":    "M4 per ounce of gold (1967 = 100)",
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

// M4GoldDownloadsHandler renders the JSON/CSV download links.
func M4GoldDownloadsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam := requestM4Range(r)
	component := templates.ChartDownloads("m4-gold", rangeParam, "average", false)

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

// M4GoldDataHandler returns the monthly ratios as JSON.
func M4GoldDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam := requestM4Range(r)
	chartData, err := getM4Gold(rangeParam)
	if err != nil {
		log.Print("failed to get m4-gold data:", err)
		http.Error(w, "Unable to load chart data. Please try again later.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"m4-gold-data.json\"")
	if err := json.NewEncoder(w).Encode(chartData); err != nil {
		log.Print("failed to encode JSON:", err)
	}
}

// M4GoldCSVHandler returns the monthly ratios as CSV.
func M4GoldCSVHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam := requestM4Range(r)
	chartData, err := getM4Gold(rangeParam)
	if err != nil {
		log.Print("failed to get m4-gold data:", err)
		http.Error(w, "Unable to load chart data. Please try again later.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"m4-gold-data.csv\"")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	if err := writer.Write([]string{"date", "m4_to_gold"}); err != nil {
		log.Print("failed to write CSV header:", err)
		return
	}
	for _, d := range chartData {
		row := []string{d.Date, strings.TrimRight(fmt.Sprintf("%.2f", d.Value), "0")}
		if err := writer.Write(row); err != nil {
			log.Print("failed to write CSV row:", err)
			return
		}
	}
}
