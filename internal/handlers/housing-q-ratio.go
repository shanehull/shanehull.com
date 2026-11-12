package handlers

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/shanehull/shanehull.com/internal/abs"
	"github.com/shanehull/shanehull.com/internal/cache"
	"github.com/shanehull/shanehull.com/internal/charts"
	"github.com/shanehull/shanehull.com/internal/rba"
	"github.com/shanehull/shanehull.com/internal/templates"
)

// Housing Q ratio: the market value of the dwelling stock, net of the housing
// credit written against it, over the cost of replacing the structures. The
// ratio is detrended by default, dividing each reading by the running geometric
// mean of every reading to that date.

const (
	qRatioCacheKey = "housing-q-ratio"
	qRatioTTL      = 24 * time.Hour
)

var qRatioCache = cache.New()

// qPoint is one observation of the ratio.
type qPoint struct {
	Date  time.Time
	Value float64
}

func loadQRatio(ctx context.Context) ([]qPoint, error) {
	if cached, found := qRatioCache.Get(qRatioCacheKey); found {
		return cached.([]qPoint), nil
	}

	fetchCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	raw, err := fetchQRatio(fetchCtx)
	if err != nil {
		return nil, err
	}
	qRatioCache.Set(qRatioCacheKey, raw, qRatioTTL)
	return raw, nil
}

// fetchQRatio fetches the ABS household balance sheet and RBA housing credit
// concurrently, then builds the raw ratio.
func fetchQRatio(ctx context.Context) ([]qPoint, error) {
	var (
		housing, replacement, credit []qPoint
		errSheet, errCredit          error
	)

	var wg sync.WaitGroup
	wg.Go(func() {
		value, dwellings, err := abs.New().FetchHouseholdBalanceSheet(ctx)
		if err != nil {
			errSheet = err
			return
		}
		for _, p := range value {
			housing = append(housing, qPoint{Date: p.Date, Value: p.Value})
		}
		for _, p := range dwellings {
			replacement = append(replacement, qPoint{Date: p.Date, Value: p.Value})
		}
	})
	wg.Go(func() {
		points, err := rba.New().FetchHousingCredit(ctx)
		if err != nil {
			errCredit = err
			return
		}
		for _, p := range points {
			credit = append(credit, qPoint{Date: p.Date, Value: p.Value})
		}
	})
	wg.Wait()

	for _, err := range []error{errSheet, errCredit} {
		if err != nil {
			return nil, err
		}
	}

	var equity []qPoint
	for _, p := range housing {
		debt, ok := qAsOf(credit, p.Date)
		if !ok {
			continue
		}
		equity = append(equity, qPoint{Date: p.Date, Value: p.Value - debt})
	}

	ratio := qRatio(equity, replacement)
	if len(ratio) < 8 {
		return nil, fmt.Errorf("q ratio: too few observations")
	}
	return ratio, nil
}

// qRatio returns equity/replacement with replacement forward-filled onto each
// balance sheet date.
func qRatio(equity, replacement []qPoint) []qPoint {
	out := make([]qPoint, 0, len(equity))
	for _, p := range equity {
		r, ok := qAsOf(replacement, p.Date)
		if !ok || r == 0 {
			continue
		}
		out = append(out, qPoint{Date: p.Date, Value: p.Value / r})
	}
	return out
}

func qAsOf(points []qPoint, t time.Time) (float64, bool) {
	idx := sort.Search(len(points), func(i int) bool { return points[i].Date.After(t) })
	if idx == 0 {
		return 0, false
	}
	return points[idx-1].Value, true
}

// qDetrend divides each reading by the running geometric mean of every reading
// to that date. Detrending the full history before any range filter keeps a
// given quarter's value stable as the range changes.
func qDetrend(points []qPoint) []qPoint {
	out := make([]qPoint, 0, len(points))
	product := 1.0
	n := 0
	for _, p := range points {
		if p.Value == 0 {
			continue
		}
		n++
		product *= p.Value
		out = append(out, qPoint{
			Date:  p.Date,
			Value: p.Value / math.Pow(product, 1/float64(n)),
		})
	}
	return out
}

func qRatioControls(r *http.Request) (rangeParam string, detrended, showQuartiles bool) {
	rangeParam = r.URL.Query().Get("range")
	if rangeParam == "" {
		rangeParam = "max"
	}

	// The checkbox and its off fallback both post under "detrend", so treat any
	// "on" as detrended regardless of parameter order.
	values := r.URL.Query()["detrend"]
	detrended = len(values) == 0
	for _, value := range values {
		if value == "on" {
			detrended = true
			break
		}
	}
	return rangeParam, detrended, r.URL.Query().Get("quartiles") == "on"
}

// qRatioSeries converts the stored ratio into chart rows for the selected range
// and detrending choice. Quartiles are the 25th and 75th percentiles of the
// displayed range.
func qRatioSeries(points []qPoint, rangeParam string, detrended, showQuartiles bool) []templates.LineChartData {
	if detrended {
		points = qDetrend(points)
	}

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

// HousingQRatioHandler renders the housing Q ratio chart.
func HousingQRatioHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam, detrended, showQuartiles := qRatioControls(r)
	points, err := loadQRatio(r.Context())
	if err != nil {
		log.Print("failed to get housing q ratio data:", err)
		renderError(w, "Unable to load chart data. Please try again later.")
		return
	}

	data := qRatioSeries(points, rangeParam, detrended, showQuartiles)
	if len(data) == 0 {
		renderError(w, "No Q ratio data in the selected time range.")
		return
	}

	options := map[string]string{
		"mainLabel":     "Q ratio, detrended",
		"yAxisLabel":    "Ratio to running geometric mean",
		"showQuartiles": "false",
		"showAverage":   "false",
	}
	if !detrended {
		options["mainLabel"] = "Q ratio, raw"
		options["yAxisLabel"] = "Equity / replacement cost"
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

// HousingQRatioDownloadsHandler renders the JSON and CSV download links.
func HousingQRatioDownloadsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam, detrended, showQuartiles := qRatioControls(r)
	component := templates.ChartDownloads("housing-q-ratio", rangeParam, map[string]string{
		"detrend":   templates.Param(detrended),
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

// HousingQRatioDataHandler returns the ratio as JSON.
func HousingQRatioDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam, detrended, showQuartiles := qRatioControls(r)
	points, err := loadQRatio(r.Context())
	if err != nil {
		log.Print("failed to get housing q ratio data:", err)
		http.Error(w, "Unable to load chart data. Please try again later.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"housing-q-ratio-data.json\"")
	if err := json.NewEncoder(w).Encode(qRatioSeries(points, rangeParam, detrended, showQuartiles)); err != nil {
		log.Print("failed to encode JSON:", err)
	}
}

// HousingQRatioCSVHandler returns the ratio as CSV.
func HousingQRatioCSVHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam, detrended, showQuartiles := qRatioControls(r)
	points, err := loadQRatio(r.Context())
	if err != nil {
		log.Print("failed to get housing q ratio data:", err)
		http.Error(w, "Unable to load chart data. Please try again later.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"housing-q-ratio-data.csv\"")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	header := []string{"date", "q"}
	if showQuartiles {
		header = append(header, "quartile1", "quartile3")
	}
	if err := writer.Write(header); err != nil {
		log.Print("failed to write CSV header:", err)
		return
	}
	for _, row := range qRatioSeries(points, rangeParam, detrended, showQuartiles) {
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
