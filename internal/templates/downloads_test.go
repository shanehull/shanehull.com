package templates

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

func renderDownloads(t *testing.T, component templ.Component) string {
	t.Helper()

	buf := new(bytes.Buffer)
	if err := component.Render(context.Background(), buf); err != nil {
		t.Fatalf("render failed: %v", err)
	}
	return buf.String()
}

func TestChartDownloadsCarriesAllParams(t *testing.T) {
	out := renderDownloads(t, ChartDownloads("demo", "10y", map[string]string{
		"average": Param(true),
		"debt":    Param(false),
	}))

	for _, want := range []string{
		`/demo/data?average=on&amp;debt=off&amp;range=10y`,
		`/demo/data.csv?average=on&amp;debt=off&amp;range=10y`,
		`download="demo-data.json"`,
		`download="demo-data.csv"`,
		`hx-boost="false"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestChartDownloadsWithoutParams(t *testing.T) {
	out := renderDownloads(t, ChartDownloads("demo", "max", nil))

	for _, want := range []string{
		`/demo/data?range=max`,
		`/demo/data.csv?range=max`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestParam(t *testing.T) {
	if got := Param(true); got != "on" {
		t.Errorf("Param(true) = %q, want on", got)
	}
	if got := Param(false); got != "off" {
		t.Errorf("Param(false) = %q, want off", got)
	}
}
