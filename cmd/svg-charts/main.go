// Command svg-charts renders SVG figures from a JSON chart spec.
//
// Usage:
//
//	go run ./cmd/svg-charts content/blog/data/housing-valuation-2026-q2/figures.json
package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type spec struct {
	Figures []figure `json:"figures"`
}

type figure struct {
	Type      string           `json:"type"`
	Output    string           `json:"output"`
	Title     string           `json:"title"`
	Subtitle  string           `json:"subtitle"`
	Source    string           `json:"source"`
	Data      string           `json:"data"`
	Start     string           `json:"start"`
	Quartiles bool             `json:"quartiles"`
	Panels    []panel          `json:"panels"`
	Labels    map[string]label `json:"labels"`
}

type panel struct {
	Title  string `json:"title"`
	Field  string `json:"field"`
	Format string `json:"format"`
}

type label struct {
	Label  string `json:"label"`
	Detail string `json:"detail"`
	Accent bool   `json:"accent"`
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: svg-charts <spec.json>")
	}
	specPath := os.Args[1]
	body, err := os.ReadFile(specPath)
	if err != nil {
		log.Fatal(err)
	}
	var s spec
	if err := json.Unmarshal(body, &s); err != nil {
		log.Fatal(err)
	}

	baseDir := filepath.Dir(specPath)
	for _, fig := range s.Figures {
		var svg string
		switch fig.Type {
		case "series":
			svg, err = seriesSVG(fig, baseDir)
		case "lollipop":
			svg, err = lollipopSVG(fig, baseDir)
		default:
			log.Fatalf("unknown figure type %q", fig.Type)
		}
		if err != nil {
			log.Fatal(err)
		}
		if err := os.WriteFile(fig.Output, []byte(svg), 0o644); err != nil {
			log.Fatal(err)
		}
		fmt.Println("wrote", fig.Output)
	}
}

func seriesSVG(fig figure, baseDir string) (string, error) {
	rows, err := readCSV(filepath.Join(baseDir, fig.Data))
	if err != nil {
		return "", err
	}

	const (
		width       = 900
		left        = 78
		right       = 132
		top         = 118
		bottom      = 72
		panelHeight = 186
		panelGap    = 86
	)
	plotWidth := width - left - right
	height := top + len(fig.Panels)*panelHeight + (len(fig.Panels)-1)*panelGap + bottom

	startTime := time.Time{}
	if fig.Start != "" {
		var err error
		startTime, err = time.Parse("2006-01-02", fig.Start)
		if err != nil {
			return "", fmt.Errorf("parse start %q: %w", fig.Start, err)
		}
	}

	times := make([]time.Time, 0, len(rows))
	for _, row := range rows {
		t, err := time.Parse("2006-01-02", row["date"])
		if err != nil {
			return "", fmt.Errorf("parse date %q: %w", row["date"], err)
		}
		if !startTime.IsZero() && t.Before(startTime) {
			continue
		}
		times = append(times, t)
	}
	if len(times) == 0 {
		return "", fmt.Errorf("no rows on or after the start date")
	}
	minTime, maxTime := times[0], times[0]
	for _, t := range times {
		if t.Before(minTime) {
			minTime = t
		}
		if t.After(maxTime) {
			maxTime = t
		}
	}
	span := maxTime.Sub(minTime).Hours()
	x := func(t time.Time) float64 {
		return float64(left) + (t.Sub(minTime).Hours()/span)*float64(plotWidth)
	}

	startYear, endYear := minTime.Year(), maxTime.Year()
	var yearTicks []int
	for year := int(math.Ceil(float64(startYear)/10) * 10); year <= endYear; year += 10 {
		yearTicks = append(yearTicks, year)
	}
	if len(yearTicks) == 0 || yearTicks[0] != startYear {
		yearTicks = append([]int{startYear}, yearTicks...)
	}
	if yearTicks[len(yearTicks)-1] != endYear {
		yearTicks = append(yearTicks, endYear)
	}

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" role="img" aria-labelledby="title desc">
  <title id="title">%s</title>
  <desc id="desc">%s</desc>
  <style>%s</style>

  <rect width="%d" height="%d" class="bg"/>
  <text x="%d" y="38" class="heading">%s</text>
  <text x="%d" y="62" class="subhead">%s</text>
`, width, height, width, height, escape(fig.Title), escape(fig.Subtitle), styleBlock, width, height, left, escape(fig.Title), left, escape(fig.Subtitle))

	for i, p := range fig.Panels {
		topY := top + i*(panelHeight+panelGap)
		bottomY := topY + panelHeight

		var points []time.Time
		var values []float64
		for _, row := range rows {
			raw := row[p.Field]
			if raw == "" {
				continue
			}
			t, err := time.Parse("2006-01-02", row["date"])
			if err != nil {
				return "", err
			}
			if !startTime.IsZero() && t.Before(startTime) {
				continue
			}
			v, err := strconv.ParseFloat(raw, 64)
			if err != nil {
				return "", fmt.Errorf("parse %s %q: %w", p.Field, raw, err)
			}
			points = append(points, t)
			values = append(values, v)
		}
		if len(values) == 0 {
			return "", fmt.Errorf("panel %q has no data", p.Title)
		}

		minimum, maximum := values[0], values[0]
		for _, v := range values {
			if v < minimum {
				minimum = v
			}
			if v > maximum {
				maximum = v
			}
		}
		if maximum == minimum {
			maximum += 0.5
			minimum -= 0.5
		}
		pad := (maximum - minimum) * 0.08
		minimum -= pad
		maximum += pad
		y := func(v float64) float64 {
			return float64(bottomY) - ((v-minimum)/(maximum-minimum))*panelHeight
		}

		var path strings.Builder
		for j, v := range values {
			cmd := "L"
			if j == 0 {
				cmd = "M"
			}
			fmt.Fprintf(&path, "%s%.1f,%.1f ", cmd, x(points[j]), y(v))
		}

		fmt.Fprintf(&b, `<g>
    <text x="%d" y="%d" class="panel-title">%s</text>
    <text x="%d" y="%d" class="latest" text-anchor="end">%s</text>

    <rect x="%d" y="%d" width="%d" height="%d" class="plot"/>
`, left, topY-30, escape(p.Title), left+plotWidth, topY-30, fmt.Sprintf(p.Format, values[len(values)-1]), left, topY, plotWidth, panelHeight)

		for _, fraction := range []float64{0, 0.5, 1} {
			value := minimum + (maximum-minimum)*fraction
			py := y(value)
			fmt.Fprintf(&b, `    <line x1="%d" y1="%.1f" x2="%d" y2="%.1f" class="grid"/>
    <text x="%d" y="%.1f" class="tick" text-anchor="end">%s</text>
`, left, py, left+plotWidth, py, left-10, py+4, fmt.Sprintf(p.Format, value))
		}
		for _, year := range yearTicks {
			tickTime := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
			fmt.Fprintf(&b, `    <text x="%.1f" y="%d" class="tick" text-anchor="middle">%d</text>
`, x(tickTime), bottomY+22, year)
		}

		if fig.Quartiles {
			q1 := quantile(values, 0.25)
			q3 := quantile(values, 0.75)
			fmt.Fprintf(&b, `    <line x1="%d" y1="%.1f" x2="%d" y2="%.1f" class="quartile q3"/>
    <line x1="%d" y1="%.1f" x2="%d" y2="%.1f" class="quartile q1"/>
    <text x="%d" y="%.1f" class="quartile-label q3">75th</text>
    <text x="%d" y="%.1f" class="quartile-label q1">25th</text>
`, left, y(q3), left+plotWidth, y(q3), left, y(q1), left+plotWidth, y(q1), left+plotWidth+10, y(q3)+4, left+plotWidth+10, y(q1)+4)
		}

		fmt.Fprintf(&b, `    <path d="%s" class="line"/>
`, path.String())

		fmt.Fprintf(&b, `  </g>
`)
	}

	fmt.Fprintf(&b, `
  <text x="%d" y="%d" class="foot">%s</text>
</svg>
`, left, height-22, escape(fig.Source))
	return b.String(), nil
}

func lollipopSVG(fig figure, baseDir string) (string, error) {
	rows, err := readCSV(filepath.Join(baseDir, fig.Data))
	if err != nil {
		return "", err
	}

	type bar struct {
		label  string
		detail string
		move   float64
		accent bool
	}
	bars := make([]bar, 0, len(rows))
	for _, row := range rows {
		value, err := strconv.ParseFloat(row["value"], 64)
		if err != nil {
			return "", fmt.Errorf("parse value %q: %w", row["value"], err)
		}
		move, err := strconv.ParseFloat(row["move_required_pct"], 64)
		if err != nil {
			return "", fmt.Errorf("parse move %q: %w", row["move_required_pct"], err)
		}
		l, ok := fig.Labels[row["measure"]+"|"+row["destination"]]
		if !ok {
			l = fig.Labels[row["measure"]]
		}
		detail := l.Detail
		if detail == "" {
			detail = row["destination"]
		} else {
			detail = fmt.Sprintf(detail, value)
		}
		bars = append(bars, bar{label: l.Label, detail: detail, move: move / 100, accent: l.Accent})
	}
	sort.SliceStable(bars, func(i, j int) bool { return bars[i].move > bars[j].move })

	const (
		width  = 900
		left   = 258
		right  = 112
		top    = 116
		rowH   = 56
		bottom = 70
	)
	plotW := width - left - right
	height := top + len(bars)*rowH + bottom
	maxFall := 0.0
	for _, b := range bars {
		if b.move > maxFall {
			maxFall = b.move
		}
	}
	axisMax := math.Ceil(maxFall*10)/10 + 0.05
	x := func(move float64) float64 {
		return float64(left) + (move/axisMax)*float64(plotW)
	}

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" role="img" aria-labelledby="title desc">
  <title id="title">%s</title>
  <desc id="desc">%s</desc>
  <style>%s</style>

  <rect width="%d" height="%d" class="bg"/>
  <text x="42" y="40" class="heading">%s</text>
  <text x="42" y="66" class="subhead">%s</text>
`, width, height, width, height, escape(fig.Title), escape(fig.Subtitle), styleBlock, width, height, escape(fig.Title), escape(fig.Subtitle))

	for tick := 0.0; tick <= axisMax; tick += 0.1 {
		fmt.Fprintf(&b, `  <line x1="%.1f" y1="%d" x2="%.1f" y2="%d" class="grid"/>
  <text x="%.1f" y="%d" class="tick" text-anchor="middle">%d%%</text>
`, x(tick), top-10, x(tick), top+len(bars)*rowH+2, x(tick), top+len(bars)*rowH+26, int(math.Round(tick*100)))
	}

	for i, row := range bars {
		y := top + i*rowH + rowH/2
		end := x(row.move)
		accentClass := ""
		if row.accent {
			accentClass = " accent"
		}
		fmt.Fprintf(&b, `  <text x="%d" y="%d" class="row-label" text-anchor="end">%s</text>
  <text x="%d" y="%d" class="row-detail" text-anchor="end">%s</text>
  <line x1="%.1f" y1="%d" x2="%.1f" y2="%d" class="stem%s"/>
  <circle cx="%.1f" cy="%d" r="6" class="dot%s"/>
  <text x="%.1f" y="%d" class="fall">%d%%</text>
`, left-18, y-2, escape(row.label), left-18, y+18, escape(row.detail), x(0), y, end, y, accentClass, end, y, accentClass, end+14, y+4, int(math.Round(row.move*100)))
	}

	fmt.Fprintf(&b, `
  <text x="42" y="%d" class="subhead">%s</text>
</svg>
`, height-18, escape(fig.Source))
	return b.String(), nil
}

func readCSV(path string) ([]map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, err
	}
	out := make([]map[string]string, 0, len(rows)-1)
	for _, row := range rows[1:] {
		record := make(map[string]string, len(rows[0]))
		for i, key := range rows[0] {
			if i < len(row) {
				record[key] = row[i]
			}
		}
		out = append(out, record)
	}
	return out, nil
}

func quantile(values []float64, p float64) float64 {
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	index := p * float64(len(sorted)-1)
	lo := int(math.Floor(index))
	hi := int(math.Ceil(index))
	return sorted[lo]*(1-(index-float64(lo))) + sorted[hi]*(index-float64(lo))
}

func escape(value string) string {
	replacer := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&apos;")
	return replacer.Replace(value)
}

const styleBlock = `text { font-family: ui-sans-serif, system-ui, -apple-system, "Segoe UI", sans-serif; }
    .bg { fill: #f5f2ed; }
    .heading { fill: #3a3432; font-size: 14px; letter-spacing: 0.08em; text-transform: uppercase; }
    .subhead { fill: #6b7280; font-size: 13px; }
    .panel-title { fill: #3a3432; font-size: 15px; font-weight: 600; }
    .latest { fill: #3a3432; font-size: 15px; font-weight: 600; font-variant-numeric: tabular-nums; }
    .plot { fill: none; stroke: rgba(75, 85, 99, 0.18); }
    .grid { stroke: rgba(75, 85, 99, 0.15); stroke-width: 1; }
    .tick { fill: #6b7280; font-size: 12px; font-variant-numeric: tabular-nums; }
    .line { fill: none; stroke: #3b82f6; stroke-width: 2; stroke-linejoin: round; stroke-linecap: round; }
    .quartile { stroke-width: 1; stroke-dasharray: 5 5; }
    .quartile.q1 { stroke: #c86464; }
    .quartile.q3 { stroke: #64a864; }
    .quartile-label { font-size: 12px; }
    .quartile-label.q1 { fill: #c86464; }
    .quartile-label.q3 { fill: #64a864; }
    .row-label { fill: #3a3432; font-size: 14px; font-weight: 600; }
    .row-detail { fill: #6b7280; font-size: 12px; }
    .stem { stroke: #9aa6b8; stroke-width: 2; }
    .stem.accent { stroke: #b86e3b; }
    .dot { fill: #3b82f6; }
    .dot.accent { fill: #b86e3b; }
    .fall { fill: #3a3432; font-size: 13px; font-weight: 600; font-variant-numeric: tabular-nums; }
    .foot { fill: #6b7280; font-size: 12px; }

    @media (prefers-color-scheme: dark) {
      .bg { fill: #1a1b1e; }
      .heading, .panel-title, .latest, .row-label, .fall { fill: #e5dfd7; }
      .subhead, .tick, .row-detail, .foot { fill: #9ca3af; }
      .plot { stroke: rgba(209, 213, 219, 0.2); }
      .grid { stroke: rgba(209, 213, 219, 0.15); }
      .quartile.q1 { stroke: #f87171; }
      .quartile.q3 { stroke: #4ade80; }
      .quartile-label.q1 { fill: #f87171; }
      .quartile-label.q3 { fill: #4ade80; }
      .stem { stroke: #6b7280; }
      .stem.accent { stroke: #cc8850; }
      .dot.accent { fill: #cc8850; }
    }`
