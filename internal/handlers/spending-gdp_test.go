package handlers

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/shanehull/shanehull.com/internal/templates"
)

func TestSpendCountriesConfigured(t *testing.T) {
	want := map[string]string{
		"us": "USA",
		"gb": "GBR",
		"fr": "FRA",
		"de": "DEU",
		"jp": "JPN",
		"au": "AUS",
		"in": "IND",
	}
	if len(spendCountries) != len(want) {
		t.Fatalf("expected %d countries, got %d", len(want), len(spendCountries))
	}
	for _, c := range spendCountries {
		if got := want[c.Code]; got != c.WBISO {
			t.Errorf("WB code for %s = %s, want %s", c.Code, c.WBISO, got)
		}
		if c.Name == "" || c.Color == "" {
			t.Errorf("country %s missing name or color", c.Code)
		}
	}
}

func TestSpendingDownloadsRenders(t *testing.T) {
	component := templates.MultiChartDownloads("spending-gdp", "50y", []string{"us", "fr"})

	buf := new(bytes.Buffer)
	if err := component.Render(context.Background(), buf); err != nil {
		t.Fatalf("render failed: %v", err)
	}

	out := buf.String()
	for _, want := range []string{
		`/spending-gdp/data?range=50y&amp;us=on&amp;fr=on`,
		`/spending-gdp/data.csv?range=50y&amp;us=on&amp;fr=on`,
		`spending-gdp-data.json`,
		`spending-gdp-data.csv`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("downloads render missing %q:\n%s", want, out)
		}
	}
}