package handlers

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/shanehull/shanehull.com/internal/templates"
)

const (
	equityID    = "NCBCEL"
	networthID  = "TNWMVBSNNCB"
	fredBaseURL = "https://api.stlouisfed.org/fred/series/observations"
)

type FinancialData struct {
	Date     time.Time
	Equity   float64
	NetWorth float64
	MSIndex  float64
}

type FredObservation struct {
	Date  string `json:"date"`
	Value string `json:"value"`
}

type FredResponse struct {
	Observations []FredObservation `json:"observations"`
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

	component := templates.ChartDownloads("msindex", rangeParam, showQuartiles)

	buf := new(bytes.Buffer)
	defer buf.Reset()

	if err := component.Render(r.Context(), buf); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
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

	startDate := calculateStartDate(rangeParam)

	equity, err := fetchFredData(equityID, startDate)
	if err != nil {
		log.Print("failed to fetch equity data:", err)
		http.Error(w, "Unable to load chart data. Please try again later.", http.StatusInternalServerError)
		return
	}

	networth, err := fetchFredData(networthID, startDate)
	if err != nil {
		log.Print("failed to fetch networth data:", err)
		http.Error(w, "Unable to load chart data. Please try again later.", http.StatusInternalServerError)
		return
	}

	data := mergeAndCalculate(equity, networth)
	if len(data) == 0 {
		http.Error(w, "No data available for the selected time range.", http.StatusInternalServerError)
		return
	}

	var q1, q3 []float64
	if showQuartiles {
		q1, q3 = calculateQuartiles(data)
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

	for i, d := range data {
		q1Val := 0.0
		q3Val := 0.0
		if showQuartiles && i < len(q1) && i < len(q3) {
			q1Val = q1[i]
			q3Val = q3[i]
		}

		row := []string{
			d.Date.Format("2006-01-02"),
			fmt.Sprintf("%.6f", d.MSIndex),
		}

		if showQuartiles {
			row = append(row, fmt.Sprintf("%.6f", q1Val), fmt.Sprintf("%.6f", q3Val))
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

	startDate := calculateStartDate(rangeParam)

	equity, err := fetchFredData(equityID, startDate)
	if err != nil {
		log.Print("failed to fetch equity data:", err)
		http.Error(w, "Unable to load chart data. Please try again later.", http.StatusInternalServerError)
		return
	}

	networth, err := fetchFredData(networthID, startDate)
	if err != nil {
		log.Print("failed to fetch networth data:", err)
		http.Error(w, "Unable to load chart data. Please try again later.", http.StatusInternalServerError)
		return
	}

	data := mergeAndCalculate(equity, networth)
	if len(data) == 0 {
		http.Error(w, "No data available for the selected time range.", http.StatusInternalServerError)
		return
	}

	var q1, q3 []float64
	if showQuartiles {
		q1, q3 = calculateQuartiles(data)
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

	startDate := calculateStartDate(rangeParam)

	equity, err := fetchFredData(equityID, startDate)
	if err != nil {
		log.Print("failed to fetch equity data:", err)
		renderError(w, "Unable to load chart data. Please try again later.")
		return
	}

	networth, err := fetchFredData(networthID, startDate)
	if err != nil {
		log.Print("failed to fetch networth data:", err)
		renderError(w, "Unable to load chart data. Please try again later.")
		return
	}

	data := mergeAndCalculate(equity, networth)
	if len(data) == 0 {
		renderError(w, "No data available for the selected time range.")
		return
	}

	var q1, q3 []float64
	if showQuartiles {
		q1, q3 = calculateQuartiles(data)
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

	options := map[string]string{
		"mainLabel":  "Misesian Stationary Index",
		"yAxisLabel": "Index Value",
	}
	component := templates.LineChart("chart-canvas", chartData, showQuartiles, options)

	buf := new(bytes.Buffer)
	defer buf.Reset()

	if err := component.Render(r.Context(), buf); err != nil {
		log.Print("failed to render component:", err)
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

func fetchFredData(seriesID string, startDate time.Time) ([]FinancialData, error) {
	apiKey := os.Getenv("FRED_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("FRED_API_KEY environment variable not set")
	}

	startDateStr := startDate.Format("2006-01-02")
	url := fredBaseURL + "?series_id=" + seriesID + "&api_key=" + apiKey + "&file_type=json&observation_start=" + startDateStr + "&frequency=q&units=lin"

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var fredResp FredResponse
	if err := json.Unmarshal(body, &fredResp); err != nil {
		return nil, err
	}

	data := make([]FinancialData, 0)
	for _, obs := range fredResp.Observations {
		if obs.Value == "." {
			continue
		}

		parsedDate, err := time.Parse("2006-01-02", obs.Date)
		if err != nil {
			continue
		}

		value, err := strconv.ParseFloat(obs.Value, 64)
		if err != nil {
			continue
		}

		data = append(data, FinancialData{
			Date: parsedDate,
		})

		if seriesID == equityID {
			data[len(data)-1].Equity = value
		} else {
			data[len(data)-1].NetWorth = value
		}
	}

	return data, nil
}

func mergeAndCalculate(equity, networth []FinancialData) []FinancialData {
	networthMap := make(map[string]float64)
	for _, nw := range networth {
		networthMap[nw.Date.Format("2006-01-02")] = nw.NetWorth
	}

	merged := make([]FinancialData, 0)
	for _, eq := range equity {
		dateStr := eq.Date.Format("2006-01-02")
		if nwValue, ok := networthMap[dateStr]; ok {
			merged = append(merged, FinancialData{
				Date:     eq.Date,
				Equity:   eq.Equity,
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

func calculateQuartiles(data []FinancialData) ([]float64, []float64) {
	if len(data) == 0 {
		return nil, nil
	}

	values := make([]float64, len(data))
	for i, d := range data {
		values[i] = d.MSIndex
	}

	sortedValues := make([]float64, len(values))
	copy(sortedValues, values)
	sort.Float64s(sortedValues)

	q1Val := percentile(sortedValues, 0.25)
	q3Val := percentile(sortedValues, 0.75)

	q1Slice := make([]float64, len(data))
	q3Slice := make([]float64, len(data))
	for i := range data {
		q1Slice[i] = q1Val
		q3Slice[i] = q3Val
	}

	return q1Slice, q3Slice
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if len(sorted) == 1 {
		return sorted[0]
	}

	index := p * float64(len(sorted)-1)
	lower := int(index)
	upper := lower + 1
	weight := index - float64(lower)

	if upper >= len(sorted) {
		return sorted[lower]
	}

	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

func calculateStartDate(rangeParam string) time.Time {
	now := time.Now()
	switch rangeParam {
	case "1y":
		return now.AddDate(-1, 0, 0)
	case "5y":
		return now.AddDate(-5, 0, 0)
	case "10y":
		return now.AddDate(-10, 0, 0)
	case "20y":
		return now.AddDate(-20, 0, 0)
	case "50y":
		return now.AddDate(-50, 0, 0)
	default: // "max"
		return time.Date(1952, 1, 1, 0, 0, 0, 0, time.UTC)
	}
}

func renderError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := `<div class="chart-error">
		<strong>Error:</strong> ` + message + `
	</div>`
	w.Write([]byte(html))
}
