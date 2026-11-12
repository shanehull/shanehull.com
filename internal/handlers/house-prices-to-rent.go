package handlers

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/shanehull/shanehull.com/internal/cache"
	"github.com/shanehull/shanehull.com/internal/charts"
	"github.com/shanehull/shanehull.com/internal/oecd"
	"github.com/shanehull/shanehull.com/internal/templates"
)

// House prices to rent: the OECD nominal house price to rent ratio for
// Australia, quarterly and oldest first.

const (
	priceToRentCacheKey = "house-prices-to-rent"
	priceToRentTTL      = 24 * time.Hour
)

var priceToRentCache = cache.New()

// rentPoint is one observation of the ratio.
type rentPoint struct {
	Date  time.Time
	Value float64
}

func loadPriceToRent(ctx context.Context) ([]rentPoint, error) {
	if cached, found := priceToRentCache.Get(priceToRentCacheKey); found {
		return cached.([]rentPoint), nil
	}

	fetchCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	raw, err := fetchPriceToRent(fetchCtx)
	if err != nil {
		return nil, err
	}
	priceToRentCache.Set(priceToRentCacheKey, raw, priceToRentTTL)
	return raw, nil
}

func fetchPriceToRent(ctx context.Context) ([]rentPoint, error) {
	points, err := oecd.New().FetchPriceToRent(ctx, "AUS")
	if err != nil {
		return nil, err
	}
	ratio := make([]rentPoint, 0, len(points))
	for _, p := range points {
		ratio = append(ratio, rentPoint{Date: p.Date, Value: p.Value})
	}
	if len(ratio) < 8 {
		return nil, fmt.Errorf("house prices to rent: too few observations")
	}
	return ratio, nil
}

func priceToRentControls(r *http.Request) (rangeParam string, showQuartiles bool) {
	rangeParam = r.URL.Query().Get("range")
	if rangeParam == "" {
		rangeParam = "max"
	}
	return rangeParam, r.URL.Query().Get("quartiles") == "on"
}

// priceToRentSeries converts the ratio into chart rows for the selected range.
// Quartiles are the 25th and 75th percentiles of the displayed range.
func priceToRentSeries(points []rentPoint, rangeParam string, showQuartiles bool) []templates.LineChartData {
	start := charts.CalculateRangeStart(rangeParam)
	data := make([]templates.LineChartData, 0, len(points))
	for _, p := range points {
		if start != nil && p.Date.Before(*start) {
			continue
		}
		data = append(data, templates.LineChartData{
			Date:  p.Date.Format("2006-01-02"),
			Value: p.Value,
		})
	}

	if showQuartiles && len(data) > 0 {
		values := make([]float64, len(data))
		for i, d := range data {
			values[i] = d.Value
		}
		q1, q3 := calculateQuartiles(values)
		for i := range data {
			data[i].Quartile1 = q1[i]
			data[i].Quartile3 = q3[i]
		}
	}
	return data
}

// HousePricesToRentHandler renders the house prices to rent chart.
func HousePricesToRentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam, showQuartiles := priceToRentControls(r)
	points, err := loadPriceToRent(r.Context())
	if err != nil {
		log.Print("failed to get house prices to rent data:", err)
		renderError(w, "Unable to load chart data. Please try again later.")
		return
	}

	data := priceToRentSeries(points, rangeParam, showQuartiles)
	if len(data) == 0 {
		renderError(w, "No data in the selected time range.")
		return
	}

	options := map[string]string{
		"mainLabel":     "House prices to rent",
		"yAxisLabel":    "House price index / rent index",
		"showQuartiles": "false",
		"showAverage":   "false",
	}
	if showQuartiles {
		options["showQuartiles"] = "true"
	}
	component := templates.LineChart("chart-canvas", data, showQuartiles, options)

	buf := new(bytes.Buffer)
	defer buf.Reset()

	if err := component.Render(r.Context(), buf); err != nil {
		log.Print("failed to render component:", err)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("HX-Trigger", "initChartFromData")
	if _, err := w.Write(buf.Bytes()); err != nil {
		log.Print("failed to write response:", err)
	}
}

// HousePricesToRentDownloadsHandler renders the JSON and CSV download links.
func HousePricesToRentDownloadsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam, showQuartiles := priceToRentControls(r)
	component := templates.ChartDownloads("house-prices-to-rent", rangeParam, map[string]string{
		"quartiles": templates.Param(showQuartiles),
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

// HousePricesToRentDataHandler returns the ratio as JSON.
func HousePricesToRentDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam, showQuartiles := priceToRentControls(r)
	points, err := loadPriceToRent(r.Context())
	if err != nil {
		log.Print("failed to get house prices to rent data:", err)
		http.Error(w, "Unable to load chart data. Please try again later.", http.StatusInternalServerError)
		return
	}

	rows := priceToRentSeries(points, rangeParam, showQuartiles)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"house-prices-to-rent-data.json\"")
	if err := json.NewEncoder(w).Encode(rows); err != nil {
		log.Print("failed to encode JSON:", err)
	}
}

// HousePricesToRentCSVHandler returns the ratio as CSV.
func HousePricesToRentCSVHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam, showQuartiles := priceToRentControls(r)
	points, err := loadPriceToRent(r.Context())
	if err != nil {
		log.Print("failed to get house prices to rent data:", err)
		http.Error(w, "Unable to load chart data. Please try again later.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"house-prices-to-rent-data.csv\"")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	header := []string{"date", "price_to_rent"}
	if showQuartiles {
		header = append(header, "quartile1", "quartile3")
	}
	if err := writer.Write(header); err != nil {
		log.Print("failed to write CSV header:", err)
		return
	}
	for _, row := range priceToRentSeries(points, rangeParam, showQuartiles) {
		cells := []string{row.Date, fmt.Sprintf("%.6f", row.Value)}
		if showQuartiles {
			cells = append(cells, fmt.Sprintf("%.6f", row.Quartile1), fmt.Sprintf("%.6f", row.Quartile3))
		}
		if err := writer.Write(cells); err != nil {
			log.Print("failed to write CSV row:", err)
			return
		}
	}
}
