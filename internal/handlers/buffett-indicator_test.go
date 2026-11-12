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

func TestMergeAndCalculateBuffetIncludesDebt(t *testing.T) {
	quarter := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	marketCap := []fred.DataPoint{{Date: quarter, Value: 1_000_000}}
	gdp := []fred.DataPoint{{Date: quarter, Value: 20_000}}
	debt := []fred.DataPoint{{Date: quarter, Value: 500_000}}

	withDebt := mergeAndCalculateBuffet(marketCap, gdp, debt, true)
	if len(withDebt) != 1 {
		t.Fatalf("expected 1 point, got %d", len(withDebt))
	}
	if got := withDebt[0].Ratio; got != 7.5 {
		t.Errorf("ratio with debt = %v, want 7.5", got)
	}

	withoutDebt := mergeAndCalculateBuffet(marketCap, gdp, debt, false)
	if got := withoutDebt[0].Ratio; got != 5 {
		t.Errorf("ratio without debt = %v, want 5", got)
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

	withDebt := mergeAndCalculateBuffet(marketCap, gdp, debt, true)
	if len(withDebt) != 1 || !withDebt[0].Date.Equal(covered) {
		t.Fatalf("expected only the covered quarter, got %+v", withDebt)
	}

	withoutDebt := mergeAndCalculateBuffet(marketCap, gdp, debt, false)
	if len(withoutDebt) != 2 {
		t.Errorf("expected full history without debt, got %d points", len(withoutDebt))
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
