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
	"github.com/shanehull/shanehull.com/internal/charts"
	"github.com/shanehull/shanehull.com/internal/fred"
	"github.com/shanehull/shanehull.com/internal/templates"
)

// Debt-to-GDP: BIS total credit to general government as a percentage of GDP,
// the standard gauge of a sovereign's obligations against its servicing base.
// The BIS series are percent-of-GDP series in FRED, so the value is used
// directly, no currency conversion needed.

const debtGDPMemoTTL = 24 * time.Hour

var debtGdpCache = cache.New()

// gdpCountry binds a country's BIS debt series and its chart styling.
type gdpCountry struct {
	Code   string
	Name   string
	Series string
	Color  string
}

var gdpCountries = []gdpCountry{
	{Code: "us", Name: "United States", Series: "QUSGAN770A", Color: "#3b82f6"},
	{Code: "cn", Name: "China", Series: "QCNGAN770A", Color: "#ef4444"},
	{Code: "jp", Name: "Japan", Series: "QJPGAN770A", Color: "#d946ef"},
	{Code: "de", Name: "Germany", Series: "QDEGAN770A", Color: "#f59e0b"},
	{Code: "gb", Name: "United Kingdom", Series: "QGBGAN770A", Color: "#10b981"},
	{Code: "in", Name: "India", Series: "QINGAN770A", Color: "#06b6d4"},
	{Code: "au", Name: "Australia", Series: "QAUGAN770A", Color: "#8b5cf6"},
}

func findGdpCountry(code string) *gdpCountry {
	for i := range gdpCountries {
		if gdpCountries[i].Code == code {
			return &gdpCountries[i]
		}
	}
	return nil
}

// selectedCodesFrom resolves which countries to plot from the query string.
// A country is included when its code equals "on". Defaults to all countries
// when none are selected so the chart never renders empty.
func selectedCodesFrom(r *http.Request, all []string) []string {
	on := make([]string, 0, len(all))
	present := false
	for _, code := range all {
		v := r.URL.Query().Get(code)
		if v != "" {
			present = true
		}
		if v == "on" {
			on = append(on, code)
		}
	}
	if !present || len(on) == 0 {
		return all
	}
	return on
}

func selectedGdpCodes(r *http.Request) []string {
	all := make([]string, 0, len(gdpCountries))
	for _, c := range gdpCountries {
		all = append(all, c.Code)
	}
	return selectedCodesFrom(r, all)
}

// getDebtGdpMatrix fetches the quarterly debt-to-GDP ratio for the selected
// countries. Returns the sorted quarter labels and per-country value arrays
// aligned to those labels (nil entries for quarters without data).
func getDebtGdpMatrix(rangeParam string, codes []string) ([]string, map[string][]*float64, error) {
	cacheKey := "debt-gdp:" + rangeParam + ":" + strings.Join(codes, ",")
	if cached, found := debtGdpCache.Get(cacheKey); found {
		entry := cached.(*debtGdpCacheEntry)
		return entry.Labels, entry.Values, nil
	}

	rangeStart := charts.CalculateRangeStart(rangeParam)
	opts := &fred.FetchOptions{
		ObservationStart: rangeStart,
		Frequency:        "q",
		Units:            "lin",
	}

	countrySeries := make(map[string]map[string]float64, len(codes))
	quarterSet := make(map[string]time.Time)

	for _, code := range codes {
		def := findGdpCountry(code)
		if def == nil {
			continue
		}

		series, err := fred.FetchSeries(def.Series, opts)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to fetch %s for %s: %w", def.Series, def.Name, err)
		}

		ratios := make(map[string]float64)
		for _, d := range series {
			dateKey := d.Date.Format("2006-01-02")
			ratios[dateKey] = d.Value
			quarterSet[dateKey] = d.Date
		}
		countrySeries[code] = ratios
	}

	if len(quarterSet) == 0 {
		return nil, nil, fmt.Errorf("no data available for the selected time range")
	}

	keys := make([]string, 0, len(quarterSet))
	for k := range quarterSet {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return quarterSet[keys[i]].Before(quarterSet[keys[j]])
	})

	values := make(map[string][]*float64, len(codes))
	for _, code := range codes {
		ratios := countrySeries[code]
		arr := make([]*float64, len(keys))
		for i, dateKey := range keys {
			if v, ok := ratios[dateKey]; ok {
				arr[i] = &v
			}
		}
		values[code] = arr
	}

	entry := debtGdpCacheEntry{Labels: keys, Values: values}
	debtGdpCache.Set(cacheKey, &entry, debtGDPMemoTTL)

	return keys, values, nil
}

// debtGdpCacheEntry is the cache shape for getDebtGdpMatrix.
type debtGdpCacheEntry struct {
	Labels []string
	Values map[string][]*float64
}

func requestGdpRange(r *http.Request) string {
	rangeParam := r.URL.Query().Get("range")
	if rangeParam == "" {
		rangeParam = "max"
	}
	return rangeParam
}

// DebtGDPHandler renders the debt-to-GDP chart.
func DebtGDPHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam := requestGdpRange(r)
	codes := selectedGdpCodes(r)

	labels, values, err := getDebtGdpMatrix(rangeParam, codes)
	if err != nil {
		log.Print("failed to get debt-gdp data:", err)
		renderError(w, "Unable to load chart data. Please try again later.")
		return
	}

	series := make([]templates.MultiLineChartSeries, 0, len(codes))
	for _, code := range codes {
		def := findGdpCountry(code)
		if def == nil {
			continue
		}
		series = append(series, templates.MultiLineChartSeries{
			Label:  def.Name,
			Color:  def.Color,
			Values: values[code],
		})
	}

	component := templates.MultiLineChart("chart-canvas", labels, series, "Percent of GDP", true)

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

// DebtGDPDownloadsHandler renders the JSON/CSV download links for the
// currently selected range and countries.
func DebtGDPDownloadsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam := requestGdpRange(r)
	codes := selectedGdpCodes(r)

	component := templates.MultiChartDownloads("debt-gdp", rangeParam, codes)

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

// DebtGDPDataHandler returns the quarterly ratios as wide-format JSON keyed
// by country code.
func DebtGDPDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam := requestGdpRange(r)
	codes := selectedGdpCodes(r)

	labels, values, err := getDebtGdpMatrix(rangeParam, codes)
	if err != nil {
		log.Print("failed to get debt-gdp data:", err)
		http.Error(w, "Unable to load chart data. Please try again later.", http.StatusInternalServerError)
		return
	}

	rows := make([]map[string]any, 0, len(labels))
	for i, label := range labels {
		row := map[string]any{"date": label}
		for _, code := range codes {
			if i < len(values[code]) && values[code][i] != nil {
				row[code] = *values[code][i]
			}
		}
		rows = append(rows, row)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"debt-gdp-data.json\"")
	if err := json.NewEncoder(w).Encode(rows); err != nil {
		log.Print("failed to encode JSON:", err)
	}
}

// DebtGDPCSVHandler returns the quarterly ratios as a CSV with one column
// per selected country.
func DebtGDPCSVHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rangeParam := requestGdpRange(r)
	codes := selectedGdpCodes(r)

	labels, values, err := getDebtGdpMatrix(rangeParam, codes)
	if err != nil {
		log.Print("failed to get debt-gdp data:", err)
		http.Error(w, "Unable to load chart data. Please try again later.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"debt-gdp-data.csv\"")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	header := append([]string{"date"}, codes...)
	if err := writer.Write(header); err != nil {
		log.Print("failed to write CSV header:", err)
		return
	}

	for i, label := range labels {
		row := []string{label}
		for _, code := range codes {
			cell := ""
			if i < len(values[code]) && values[code][i] != nil {
				cell = fmt.Sprintf("%.2f", *values[code][i])
			}
			row = append(row, cell)
		}
		if err := writer.Write(row); err != nil {
			log.Print("failed to write CSV row:", err)
			return
		}
	}
}
