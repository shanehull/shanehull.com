package handlers

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/shanehull/shanehull.com/internal/cache"
	"github.com/shanehull/shanehull.com/internal/charts"
	"github.com/shanehull/shanehull.com/internal/templates"
	"github.com/shanehull/shanehull.com/internal/wb"
)

// Public spending as a share of GDP: general government expense (including
// transfers and interest) from the World Bank WDI indicator
// GC.XPN.TOTL.GD.ZS, the GFS measure of total government cost against nominal
// GDP. Annual. China is not shown because its GFS expense series is not
// published for this indicator.

const (
	spendIndicator = "GC.XPN.TOTL.GD.ZS"
	spendTTL       = 24 * time.Hour
)

var spendingCache = cache.New()

// spendCountry binds a country's World Bank code and its chart styling.
type spendCountry struct {
	Code  string
	Name  string
	WBISO string
	Color string
}

var spendCountries = []spendCountry{
	{Code: "us", Name: "US", WBISO: "USA", Color: "#3b82f6"},
	{Code: "gb", Name: "GB", WBISO: "GBR", Color: "#10b981"},
	{Code: "fr", Name: "FR", WBISO: "FRA", Color: "#ef4444"},
	{Code: "de", Name: "DE", WBISO: "DEU", Color: "#f59e0b"},
	{Code: "jp", Name: "JP", WBISO: "JPN", Color: "#d946ef"},
	{Code: "au", Name: "AU", WBISO: "AUS", Color: "#8b5cf6"},
	{Code: "in", Name: "IN", WBISO: "IND", Color: "#06b6d4"},
}

func spendCodes() []string {
	codes := make([]string, 0, len(spendCountries))
	for _, c := range spendCountries {
		codes = append(codes, c.Code)
	}
	return codes
}

// getSpendingMatrix fetches general government expense as a percentage of GDP
// for every configured country. Returns the sorted annual labels and
// per-country value arrays aligned to those labels (nil for missing years).
func getSpendingMatrix(rangeParam string) ([]string, map[string][]*float64, error) {
	cacheKey := "spending-gdp:" + rangeParam
	if cached, found := spendingCache.Get(cacheKey); found {
		entry := cached.(*spendCacheEntry)
		return entry.Labels, entry.Values, nil
	}

	rangeStart := charts.CalculateRangeStart(rangeParam)
	client := wb.New()

	// Fetch every country concurrently; World Bank latency dominates the cold
	// load, so parallel round trips shrink it to roughly one call.
	seriesByCode, fetchErrs := fetchAll(len(spendCountries), func(i int) ([]wb.Point, error) {
		def := spendCountries[i]
		return client.FetchIndicator(def.WBISO, spendIndicator)
	})

	countrySeries := make(map[string]map[string]float64, len(spendCountries))
	yearSet := make(map[string]time.Time)

	for i, def := range spendCountries {
		if fetchErrs[i] != nil {
			log.Print("failed to fetch spending for", def.Name, ":", fetchErrs[i])
			continue
		}

		ratios := make(map[string]float64)
		for _, d := range seriesByCode[i] {
			if rangeStart != nil && d.Date.Before(*rangeStart) {
				continue
			}
			dateKey := d.Date.Format("2006-01-02")
			ratios[dateKey] = d.Value
			yearSet[dateKey] = d.Date
		}
		countrySeries[def.Code] = ratios
	}

	if len(yearSet) == 0 {
		return nil, nil, fmt.Errorf("no data available for the selected time range")
	}

	// Plot only the years every country reports, so ragged GFS release lags
	// (Japan and India stop a year before the US) do not leave trailing gaps.
	common := make(map[string]bool)
	init := false
	for _, keys := range countrySeries {
		if !init {
			for k := range keys {
				common[k] = true
			}
			init = true
			continue
		}
		for k := range common {
			if _, ok := keys[k]; !ok {
				delete(common, k)
			}
		}
	}

	keys := make([]string, 0, len(common))
	for k := range common {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return yearSet[keys[i]].Before(yearSet[keys[j]])
	})
	if len(keys) == 0 {
		return nil, nil, fmt.Errorf("no year is reported by every country")
	}

	values := make(map[string][]*float64, len(spendCountries))
	for _, def := range spendCountries {
		ratios := countrySeries[def.Code]
		arr := make([]*float64, len(keys))
		for i, dateKey := range keys {
			if v, ok := ratios[dateKey]; ok {
				arr[i] = &v
			}
		}
		values[def.Code] = arr
	}

	entry := spendCacheEntry{Labels: keys, Values: values}
	spendingCache.Set(cacheKey, &entry, spendTTL)

	return keys, values, nil
}

// spendCacheEntry is the cache shape for getSpendingMatrix.
type spendCacheEntry struct {
	Labels []string
	Values map[string][]*float64
}

func requestSpendRange(r *http.Request) string {
	rangeParam := r.URL.Query().Get("range")
	if rangeParam == "" {
		rangeParam = "max"
	}
	return rangeParam
}

// SpendingGDPHandler renders the public spending to GDP chart.
func SpendingGDPHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam := requestSpendRange(r)
	labels, values, err := getSpendingMatrix(rangeParam)
	if err != nil {
		log.Print("failed to get spending-gdp data:", err)
		renderError(w, "Unable to load chart data. Please try again later.")
		return
	}

	series := make([]templates.MultiLineChartSeries, 0, len(spendCountries))
	for _, def := range spendCountries {
		series = append(series, templates.MultiLineChartSeries{
			Label:  def.Name,
			Color:  def.Color,
			Values: values[def.Code],
		})
	}

	component := templates.MultiLineChart("chart-canvas", labels, series, "Percent of GDP", false)

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

// SpendingGDPDownloadsHandler renders the JSON/CSV download links.
func SpendingGDPDownloadsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam := requestSpendRange(r)
	component := templates.MultiChartDownloads("spending-gdp", rangeParam, spendCodes())

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

// SpendingGDPDataHandler returns the annual values as wide-format JSON keyed
// by country code.
func SpendingGDPDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam := requestSpendRange(r)
	labels, values, err := getSpendingMatrix(rangeParam)
	if err != nil {
		log.Print("failed to get spending-gdp data:", err)
		http.Error(w, "Unable to load chart data. Please try again later.", http.StatusInternalServerError)
		return
	}

	rows := make([]map[string]any, 0, len(labels))
	for i, label := range labels {
		row := map[string]any{"date": label}
		for _, def := range spendCountries {
			if i < len(values[def.Code]) && values[def.Code][i] != nil {
				row[def.Code] = *values[def.Code][i]
			}
		}
		rows = append(rows, row)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"spending-gdp-data.json\"")
	if err := json.NewEncoder(w).Encode(rows); err != nil {
		log.Print("failed to encode JSON:", err)
	}
}

// SpendingGDPCSVHandler returns the annual values as a CSV with one column
// per country.
func SpendingGDPCSVHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam := requestSpendRange(r)
	labels, values, err := getSpendingMatrix(rangeParam)
	if err != nil {
		log.Print("failed to get spending-gdp data:", err)
		http.Error(w, "Unable to load chart data. Please try again later.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"spending-gdp-data.csv\"")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	header := append([]string{"date"}, spendCodes()...)
	if err := writer.Write(header); err != nil {
		log.Print("failed to write CSV header:", err)
		return
	}

	for i, label := range labels {
		row := []string{label}
		for _, def := range spendCountries {
			cell := ""
			if i < len(values[def.Code]) && values[def.Code][i] != nil {
				cell = fmt.Sprintf("%.2f", *values[def.Code][i])
			}
			row = append(row, cell)
		}
		if err := writer.Write(row); err != nil {
			log.Print("failed to write CSV row:", err)
			return
		}
	}
}
