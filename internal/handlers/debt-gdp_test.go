package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shanehull/shanehull.com/internal/templates"
)

func TestGDPCountriesConfigured(t *testing.T) {
	want := map[string]string{
		"us": "QUSGAN770A",
		"cn": "QCNGAN770A",
		"jp": "QJPGAN770A",
		"de": "QDEGAN770A",
		"gb": "QGBGAN770A",
		"in": "QINGAN770A",
		"au": "QAUGAN770A",
	}
	if len(gdpCountries) != len(want) {
		t.Fatalf("expected %d countries, got %d", len(want), len(gdpCountries))
	}
	for _, c := range gdpCountries {
		if got := want[c.Code]; got != c.Series {
			t.Errorf("series for %s = %s, want %s", c.Code, c.Series, got)
		}
		if c.Name == "" || c.Color == "" {
			t.Errorf("country %s missing name or color", c.Code)
		}
	}
}

func TestSelectedCodesFrom(t *testing.T) {
	all := []string{"us", "cn", "jp", "de"}

	// No params: everything on.
	if got := selectedCodesFrom(httptest.NewRequest("GET", "/debt-gdp/chart", nil), all); len(got) != 4 {
		t.Errorf("no params should select all, got %v", got)
	}

	// Explicit subset.
	req := httptest.NewRequest("GET", "/debt-gdp/chart?us=on&jp=on", nil)
	if got := selectedCodesFrom(req, all); !equalStrings(got, []string{"us", "jp"}) {
		t.Errorf("subset = %v, want [us jp]", got)
	}

	// Country explicitly off with no others on: guard keeps all so the chart
	// never renders empty.
	req = httptest.NewRequest("GET", "/debt-gdp/chart?cn=off", nil)
	if got := selectedCodesFrom(req, all); len(got) != 4 {
		t.Errorf("all-off should fall back to all, got %v", got)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestMultiLineChartRenders(t *testing.T) {
	us := 116.4
	cn := 99.3
	var nilPtr *float64
	series := []templates.MultiLineChartSeries{
		{Label: "United States", Color: "#3b82f6", Values: []*float64{&us, nilPtr}},
		{Label: "China", Color: "#ef4444", Values: []*float64{nilPtr, &cn}},
	}
	component := templates.MultiLineChart("chart-canvas", []string{"2025-10-01", "2026-01-01"}, series, "Percent of GDP", true)

	buf := new(bytes.Buffer)
	if err := component.Render(context.Background(), buf); err != nil {
		t.Fatalf("render failed: %v", err)
	}

	out := buf.String()
	start := strings.Index(out, `data-chart="`)
	if start == -1 {
		t.Fatalf("missing data-chart attribute: %s", out)
	}

	quote := out[start+len(`data-chart="`):]
	end := strings.Index(quote, `"`)
	config := strings.ReplaceAll(quote[:end], "&#34;", `"`)

	var parsed struct {
		Labels   []string `json:"labels"`
		Datasets []struct {
			Label string     `json:"label"`
			Data  []*float64 `json:"data"`
		} `json:"datasets"`
		YAxisLabel string `json:"yAxisLabel"`
		YAxisType  string `json:"yAxisType"`
	}
	if err := json.Unmarshal([]byte(config), &parsed); err != nil {
		t.Fatalf("data-chart is not valid JSON: %v", err)
	}

	if len(parsed.Labels) != 2 || parsed.Datasets[0].Label != "United States" {
		t.Errorf("unexpected chart config: %s", config)
	}
	if parsed.Datasets[0].Data[0] == nil || *parsed.Datasets[0].Data[0] != 116.4 {
		t.Errorf("unexpected US value: %v", parsed.Datasets[0].Data)
	}
	if parsed.YAxisLabel != "Percent of GDP" {
		t.Errorf("unexpected y-axis label: %s", parsed.YAxisLabel)
	}
	if parsed.YAxisType != "logarithmic" {
		t.Errorf("unexpected y-axis type: %s", parsed.YAxisType)
	}
}

func TestMultiLineChartLinearAxis(t *testing.T) {
	var v = 100.0
	series := []templates.MultiLineChartSeries{{Label: "United States", Values: []*float64{&v}}}
	component := templates.MultiLineChart("chart-canvas", []string{"2025-10-01"}, series, "Percent of GDP", false)

	buf := new(bytes.Buffer)
	if err := component.Render(context.Background(), buf); err != nil {
		t.Fatalf("render failed: %v", err)
	}
	out := buf.String()
	start := strings.Index(out, `data-chart="`)
	config := strings.ReplaceAll(out[start+12:strings.Index(out[start:], `">`)+start], "&#34;", `"`)
	if !strings.Contains(config, `"yAxisType":"linear"`) {
		t.Errorf("expected linear axis, got %s", config)
	}
}

func TestMultiChartDownloadsRenders(t *testing.T) {
	component := templates.MultiChartDownloads("debt-gdp", "max", []string{"us", "jp"})

	buf := new(bytes.Buffer)
	if err := component.Render(context.Background(), buf); err != nil {
		t.Fatalf("render failed: %v", err)
	}

	out := buf.String()
	for _, want := range []string{
		`/debt-gdp/data?range=max&amp;us=on&amp;jp=on`,
		`/debt-gdp/data.csv?range=max&amp;us=on&amp;jp=on`,
		`debt-gdp-data.json`,
		`debt-gdp-data.csv`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("downloads render missing %q:\n%s", want, out)
		}
	}
}
