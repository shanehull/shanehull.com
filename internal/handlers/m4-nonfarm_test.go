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
	m4 := []cfs.Point{
		{Date: time.Date(1967, 1, 1, 0, 0, 0, 0, time.UTC), Index: 100},
		{Date: time.Date(1967, 3, 1, 0, 0, 0, 0, time.UTC), Index: 105},
		{Date: time.Date(1967, 6, 1, 0, 0, 0, 0, time.UTC), Index: 110},
	}
	prod := []fred.DataPoint{
		{Date: time.Date(1967, 1, 1, 0, 0, 0, 0, time.UTC), Value: 2.0}, // Q1
		{Date: time.Date(1967, 4, 1, 0, 0, 0, 0, time.UTC), Value: 3.0}, // Q2
	}

	points := buildM4Nonfarm(m4, prod)
	// Q1 needs M4 at 1967-03 and 1966-12 (missing), so it is skipped; Q2 uses
	// M4 1967-06 (110) vs 1967-03 (105): ((110/105)^4 - 1) * 100.
	if len(points) != 1 {
		t.Fatalf("got %d points, want 1", len(points))
	}
	if !points[0].Date.Equal(time.Date(1967, 4, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("unexpected date %v", points[0].Date)
	}
	wantMoney := (math.Pow(110.0/105.0, 4) - 1) * 100
	if math.Abs(points[0].Money-wantMoney) > 0.01 {
		t.Errorf("M4 growth = %v, want %v", points[0].Money, wantMoney)
	}
	if points[0].Product != 3.0 {
		t.Errorf("product = %v, want 3.0", points[0].Product)
	}
	if math.Abs(points[0].Excess-(points[0].Money-3.0)) > 0.001 {
		t.Errorf("excess mismatch: %v", points[0].Excess)
	}
}

func TestSmoothExcess(t *testing.T) {
	excess := []*float64{f64(1), f64(2), f64(3), f64(4), f64(8)}
	smooth := smoothExcess(excess)

	if smooth[0] != nil || smooth[1] != nil || smooth[2] != nil {
		t.Fatalf("first three quarters should be nil, got %v %v %v", smooth[0], smooth[1], smooth[2])
	}
	if smooth[3] == nil || *smooth[3] != 2.5 { // (1+2+3+4)/4
		t.Errorf("quarter 4 = %v, want 2.5", smooth[3])
	}
	if smooth[4] == nil || *smooth[4] != 4.25 { // (2+3+4+8)/4
		t.Errorf("quarter 5 = %v, want 4.25", smooth[4])
	}
}

func f64(v float64) *float64 { return &v }

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
