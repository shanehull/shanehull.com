package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/shanehull/shanehull.com/internal/templates"
)

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
