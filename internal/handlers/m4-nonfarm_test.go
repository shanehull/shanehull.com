package handlers

import (
	"bytes"
	"context"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/shanehull/shanehull.com/internal/cfs"
	"github.com/shanehull/shanehull.com/internal/fred"
	"github.com/shanehull/shanehull.com/internal/templates"
)

func TestBuildM4Nonfarm(t *testing.T) {
	// Divisia M4 monthly: 1967-04 = 100, 1968-04 = 110 (10% YoY by 1968 Q2).
	// OPHNFB quarterly: 1967 Q2 = 80, 1968 Q2 = 88 (10% YoY).
	m4 := []cfs.Point{
		{Date: time.Date(1967, 4, 1, 0, 0, 0, 0, time.UTC), Index: 100},
		{Date: time.Date(1968, 4, 1, 0, 0, 0, 0, time.UTC), Index: 110},
	}
	prod := []fred.DataPoint{
		{Date: time.Date(1967, 1, 1, 0, 0, 0, 0, time.UTC), Value: 75}, // 1967 Q1
		{Date: time.Date(1967, 4, 1, 0, 0, 0, 0, time.UTC), Value: 80}, // 1967 Q2
		{Date: time.Date(1968, 1, 1, 0, 0, 0, 0, time.UTC), Value: 86}, // 1968 Q1
		{Date: time.Date(1968, 4, 1, 0, 0, 0, 0, time.UTC), Value: 88}, // 1968 Q2
	}

	points := buildM4Nonfarm(m4, prod)
	// Only 1968 Q2 has a productivity level four quarters earlier (1967 Q2)
	// and a M4 level one year earlier (1967-04).
	if len(points) != 1 {
		t.Fatalf("got %d points, want 1", len(points))
	}
	if !points[0].Date.Equal(time.Date(1968, 4, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("unexpected date %v", points[0].Date)
	}
	if math.Abs(points[0].Money-10.0) > 0.01 {
		t.Errorf("M4 YoY = %v, want 10", points[0].Money)
	}
	if math.Abs(points[0].Product-10.0) > 0.01 {
		t.Errorf("productivity YoY = %v, want 10", points[0].Product)
	}
	if math.Abs(points[0].Excess) > 0.01 {
		t.Errorf("excess = %v, want ~0", points[0].Excess)
	}
}

func TestM4NonfarmDownloadsRenders(t *testing.T) {
	component := templates.MultiChartDownloads("m4-nonfarm", "50y", []string{"m4", "nonfarm"})

	buf := new(bytes.Buffer)
	if err := component.Render(context.Background(), buf); err != nil {
		t.Fatalf("render failed: %v", err)
	}

	out := buf.String()
	for _, want := range []string{
		`/m4-nonfarm/data?range=50y&amp;m4=on&amp;nonfarm=on`,
		`m4-nonfarm-data.json`,
		`m4-nonfarm-data.csv`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("downloads render missing %q:\n%s", want, out)
		}
	}
}
