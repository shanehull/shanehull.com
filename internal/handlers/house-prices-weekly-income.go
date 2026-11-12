package handlers

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/shanehull/shanehull.com/internal/abs"
	"github.com/shanehull/shanehull.com/internal/cache"
	"github.com/shanehull/shanehull.com/internal/charts"
	"github.com/shanehull/shanehull.com/internal/oecd"
	"github.com/shanehull/shanehull.com/internal/templates"
)

// House prices to weekly income: the OECD nominal house price index divided by
// ABS average weekly earnings per employee. Both series are quarterly, with
// earnings forward-filled onto each house price date.

const (
	weeklyIncomeCacheKey = "house-prices-to-weekly-income"
	weeklyIncomeTTL      = 24 * time.Hour
)

var weeklyIncomeCache = cache.New()

// incomePoint is one observation of the ratio.
type incomePoint struct {
	Date  time.Time
	Value float64
}

func loadWeeklyIncomeRatio(ctx context.Context) ([]incomePoint, error) {
	if cached, found := weeklyIncomeCache.Get(weeklyIncomeCacheKey); found {
		return cached.([]incomePoint), nil
	}

	fetchCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	raw, err := fetchWeeklyIncomeRatio(fetchCtx)
	if err != nil {
		return nil, err
	}
	weeklyIncomeCache.Set(weeklyIncomeCacheKey, raw, weeklyIncomeTTL)
	return raw, nil
}

// fetchWeeklyIncomeRatio fetches house prices and earnings concurrently, then
// divides prices by forward-filled weekly earnings.
func fetchWeeklyIncomeRatio(ctx context.Context) ([]incomePoint, error) {
	var (
		prices, earnings       []incomePoint
		errPrices, errEarnings error
	)

	var wg sync.WaitGroup
	wg.Go(func() {
		points, err := oecd.New().FetchHousePriceIndex(ctx, "AUS")
		if err != nil {
			errPrices = err
			return
		}
		for _, p := range points {
			prices = append(prices, incomePoint{Date: p.Date, Value: p.Value})
		}
	})
	wg.Go(func() {
		points, err := abs.New().FetchAverageWeeklyEarnings(ctx)
		if err != nil {
			errEarnings = err
			return
		}
		for _, p := range points {
			earnings = append(earnings, incomePoint{Date: p.Date, Value: p.Value})
		}
	})
	wg.Wait()

	for _, err := range []error{errPrices, errEarnings} {
		if err != nil {
			return nil, err
		}
	}

	ratio := make([]incomePoint, 0, len(prices))
	for _, p := range prices {
		income, ok := incomeAsOf(earnings, p.Date)
		if !ok || income == 0 {
			continue
		}
		ratio = append(ratio, incomePoint{Date: p.Date, Value: p.Value / income})
	}
	if len(ratio) < 8 {
		return nil, fmt.Errorf("house prices to weekly income: too few observations")
	}
	return ratio, nil
}

func incomeAsOf(points []incomePoint, t time.Time) (float64, bool) {
	idx := sort.Search(len(points), func(i int) bool { return points[i].Date.After(t) })
	if idx == 0 {
		return 0, false
	}
	return points[idx-1].Value, true
}

func weeklyIncomeControls(r *http.Request) string {
	rangeParam := r.URL.Query().Get("range")
	if rangeParam == "" {
		rangeParam = "max"
	}
	return rangeParam
}

func weeklyIncomeSeries(points []incomePoint, rangeParam string) []templates.LineChartData {
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
	return data
}

// HousePricesWeeklyIncomeHandler renders the house prices to weekly income chart.
func HousePricesWeeklyIncomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam := weeklyIncomeControls(r)
	points, err := loadWeeklyIncomeRatio(r.Context())
	if err != nil {
		log.Print("failed to get house prices to weekly income data:", err)
		renderError(w, "Unable to load chart data. Please try again later.")
		return
	}

	data := weeklyIncomeSeries(points, rangeParam)
	if len(data) == 0 {
		renderError(w, "No data in the selected time range.")
		return
	}

	options := map[string]string{
		"mainLabel":     "House prices to weekly income",
		"yAxisLabel":    "House price index / weekly earnings ($)",
		"showQuartiles": "false",
		"showAverage":   "false",
	}
	component := templates.LineChart("chart-canvas", data, false, options)

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

// HousePricesWeeklyIncomeDownloadsHandler renders the JSON and CSV download links.
func HousePricesWeeklyIncomeDownloadsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam := weeklyIncomeControls(r)
	component := templates.ChartDownloads("house-prices-to-weekly-income", rangeParam, nil)

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

// HousePricesWeeklyIncomeDataHandler returns the ratio as JSON.
func HousePricesWeeklyIncomeDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam := weeklyIncomeControls(r)
	points, err := loadWeeklyIncomeRatio(r.Context())
	if err != nil {
		log.Print("failed to get house prices to weekly income data:", err)
		http.Error(w, "Unable to load chart data. Please try again later.", http.StatusInternalServerError)
		return
	}

	rows := weeklyIncomeSeries(points, rangeParam)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"house-prices-to-weekly-income-data.json\"")
	if err := json.NewEncoder(w).Encode(rows); err != nil {
		log.Print("failed to encode JSON:", err)
	}
}

// HousePricesWeeklyIncomeCSVHandler returns the ratio as CSV.
func HousePricesWeeklyIncomeCSVHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam := weeklyIncomeControls(r)
	points, err := loadWeeklyIncomeRatio(r.Context())
	if err != nil {
		log.Print("failed to get house prices to weekly income data:", err)
		http.Error(w, "Unable to load chart data. Please try again later.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"house-prices-to-weekly-income-data.csv\"")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	if err := writer.Write([]string{"date", "price_to_income"}); err != nil {
		log.Print("failed to write CSV header:", err)
		return
	}
	for _, row := range weeklyIncomeSeries(points, rangeParam) {
		if err := writer.Write([]string{row.Date, fmt.Sprintf("%.6f", row.Value)}); err != nil {
			log.Print("failed to write CSV row:", err)
			return
		}
	}
}
