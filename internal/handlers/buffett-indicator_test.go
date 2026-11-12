package handlers

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/shanehull/shanehull.com/internal/fred"
	"github.com/shanehull/shanehull.com/internal/templates"
)

func TestMergeAndCalculateBuffetSubtractsDebt(t *testing.T) {
	quarter := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	marketCap := []fred.DataPoint{{Date: quarter, Value: 1_000_000}}
	gdp := []fred.DataPoint{{Date: quarter, Value: 20_000}}
	debt := []fred.DataPoint{{Date: quarter, Value: 500_000}}

	withDebt := mergeAndCalculateBuffet(marketCap, gdp, debt, true)
	if len(withDebt) != 1 {
		t.Fatalf("expected 1 point, got %d", len(withDebt))
	}
	if got := withDebt[0].Ratio; got != 2.5 {
		t.Errorf("adjusted ratio = %v, want 2.5", got)
	}

	withoutDebt := mergeAndCalculateBuffet(marketCap, gdp, debt, false)
	if got := withoutDebt[0].Ratio; got != 5 {
		t.Errorf("classic ratio = %v, want 5", got)
	}
}

func TestMergeAndCalculateBuffetSkipsQuartersWithoutDebt(t *testing.T) {
	covered := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	uncovered := time.Date(1960, time.January, 1, 0, 0, 0, 0, time.UTC)
	marketCap := []fred.DataPoint{
		{Date: uncovered, Value: 1_000_000},
		{Date: covered, Value: 1_000_000},
	}
	gdp := []fred.DataPoint{
		{Date: uncovered, Value: 20_000},
		{Date: covered, Value: 20_000},
	}
	debt := []fred.DataPoint{{Date: covered, Value: 500_000}}

	// Both modes share the debt coverage window so the x-axis does not move
	// when the debt switch is toggled.
	for _, includeDebt := range []bool{true, false} {
		merged := mergeAndCalculateBuffet(marketCap, gdp, debt, includeDebt)
		if len(merged) != 1 || !merged[0].Date.Equal(covered) {
			t.Errorf("includeDebt=%v: expected only the covered quarter, got %+v", includeDebt, merged)
		}
	}
}

func TestBuffettDownloadsRendersDebtParam(t *testing.T) {
	component := templates.ChartDownloads("buffett-indicator", "max", map[string]string{
		"average": templates.Param(true),
		"debt":    templates.Param(true),
	})

	buf := new(bytes.Buffer)
	if err := component.Render(context.Background(), buf); err != nil {
		t.Fatalf("render failed: %v", err)
	}

	out := buf.String()
	for _, want := range []string{
		`/buffett-indicator/data?average=on&amp;debt=on&amp;range=max`,
		`/buffett-indicator/data.csv?average=on&amp;debt=on&amp;range=max`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("downloads render missing %q:\n%s", want, out)
		}
	}
}
