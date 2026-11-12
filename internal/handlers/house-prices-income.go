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

// House prices to household income: the OECD nominal house price index divided
// by ABS household gross disposable income per capita. Income is seasonally
// adjusted and divided by estimated resident population.

const (
	priceToIncomeCacheKey = "house-prices-to-income"
	priceToIncomeTTL      = 24 * time.Hour
)

var priceToIncomeCache = cache.New()

// incomePoint is one observation.
type incomePoint struct {
	Date  time.Time
	Value float64
}

func loadPriceToIncomeRatio(ctx context.Context) ([]incomePoint, error) {
	if cached, found := priceToIncomeCache.Get(priceToIncomeCacheKey); found {
		return cached.([]incomePoint), nil
	}

	fetchCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	raw, err := fetchPriceToIncomeRatio(fetchCtx)
	if err != nil {
		return nil, err
	}
	priceToIncomeCache.Set(priceToIncomeCacheKey, raw, priceToIncomeTTL)
	return raw, nil
}

// fetchPriceToIncomeRatio fetches prices, household disposable income and
// population concurrently, then divides prices by income per capita.
func fetchPriceToIncomeRatio(ctx context.Context) ([]incomePoint, error) {
	var (
		prices, income, population          []incomePoint
		errPrices, errIncome, errPopulation error
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
		points, err := abs.New().FetchHouseholdDisposableIncome(ctx)
		if err != nil {
			errIncome = err
			return
		}
		for _, p := range points {
			income = append(income, incomePoint{Date: p.Date, Value: p.Value})
		}
	})
	wg.Go(func() {
		points, err := abs.New().FetchPopulation(ctx)
		if err != nil {
			errPopulation = err
			return
		}
		for _, p := range points {
			population = append(population, incomePoint{Date: p.Date, Value: p.Value})
		}
	})
	wg.Wait()

	for _, err := range []error{errPrices, errIncome, errPopulation} {
		if err != nil {
			return nil, err
		}
	}

	ratio := make([]incomePoint, 0, len(prices))
	for _, p := range prices {
		disposable, ok := incomeAsOf(income, p.Date)
		if !ok || disposable == 0 {
			continue
		}
		people, ok := incomeAsOf(population, p.Date)
		if !ok || people == 0 {
			continue
		}
		// Disposable income is published in millions of dollars.
		perCapita := disposable * 1e6 / people
		ratio = append(ratio, incomePoint{Date: p.Date, Value: p.Value / perCapita})
	}
	if len(ratio) < 8 {
		return nil, fmt.Errorf("house prices to household income: too few observations")
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

func priceToIncomeControls(r *http.Request) (rangeParam string, showQuartiles bool) {
	rangeParam = r.URL.Query().Get("range")
	if rangeParam == "" {
		rangeParam = "max"
	}
	return rangeParam, r.URL.Query().Get("quartiles") == "on"
}

// priceToIncomeSeries converts the ratio into chart rows for the selected
// range. Quartiles are the 25th and 75th percentiles of the displayed range.
func priceToIncomeSeries(points []incomePoint, rangeParam string, showQuartiles bool) []templates.LineChartData {
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

// HousePricesIncomeHandler renders the house prices to household income chart.
func HousePricesIncomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam, showQuartiles := priceToIncomeControls(r)
	points, err := loadPriceToIncomeRatio(r.Context())
	if err != nil {
		log.Print("failed to get house prices to household income data:", err)
		renderError(w, "Unable to load chart data. Please try again later.")
		return
	}

	data := priceToIncomeSeries(points, rangeParam, showQuartiles)
	if len(data) == 0 {
		renderError(w, "No data in the selected time range.")
		return
	}

	options := map[string]string{
		"mainLabel":     "House prices to income",
		"yAxisLabel":    "House price index / disposable income per capita ($)",
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

// HousePricesIncomeDownloadsHandler renders the JSON and CSV download links.
func HousePricesIncomeDownloadsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam, showQuartiles := priceToIncomeControls(r)
	component := templates.ChartDownloads("house-prices-to-income", rangeParam, map[string]string{
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

// HousePricesIncomeDataHandler returns the ratio as JSON.
func HousePricesIncomeDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam, showQuartiles := priceToIncomeControls(r)
	points, err := loadPriceToIncomeRatio(r.Context())
	if err != nil {
		log.Print("failed to get house prices to household income data:", err)
		http.Error(w, "Unable to load chart data. Please try again later.", http.StatusInternalServerError)
		return
	}

	rows := priceToIncomeSeries(points, rangeParam, showQuartiles)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"house-prices-to-income-data.json\"")
	if err := json.NewEncoder(w).Encode(rows); err != nil {
		log.Print("failed to encode JSON:", err)
	}
}

// HousePricesIncomeCSVHandler returns the ratio as CSV.
func HousePricesIncomeCSVHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam, showQuartiles := priceToIncomeControls(r)
	points, err := loadPriceToIncomeRatio(r.Context())
	if err != nil {
		log.Print("failed to get house prices to household income data:", err)
		http.Error(w, "Unable to load chart data. Please try again later.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"house-prices-to-income-data.csv\"")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	header := []string{"date", "price_to_income"}
	if showQuartiles {
		header = append(header, "quartile1", "quartile3")
	}
	if err := writer.Write(header); err != nil {
		log.Print("failed to write CSV header:", err)
		return
	}
	for _, row := range priceToIncomeSeries(points, rangeParam, showQuartiles) {
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
