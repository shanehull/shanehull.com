package cfs

import (
	"archive/zip"
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// buildXLSX produces a minimal CFS-like workbook: workbook.xml, rels and one
// worksheet named "Broad" with two header rows then A=serial date, B=index.
func buildXLSX(sheetName string, serials []float64, indexes []float64) []byte {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	write := func(name, body string) {
		w, err := zw.Create(name)
		if err != nil {
			panic(err)
		}
		_, _ = w.Write([]byte(body))
	}

	write("xl/workbook.xml", fmt.Sprintf(
		`<?xml version="1.0"?><workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets><sheet name="%s" sheetId="1" r:id="rId1"/></sheets></workbook>`,
		sheetName))

	write("xl/_rels/workbook.xml.rels",
		`<?xml version="1.0"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/></Relationships>`)

	var sheet strings.Builder
	sheet.WriteString(`<?xml version="1.0"?><worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>`)
	sheet.WriteString(`<row r="1"><c r="A1"><v>0</v></c></row><row r="2"><c r="A2"><v>0</v></c></row>`)
	for i := range serials {
		row := i + 3
		fmt.Fprintf(&sheet, `<row r="%d"><c r="A%d"><v>%f</v></c><c r="B%d"><v>%f</v></c></row>`, row, row, serials[i], row, indexes[i])
	}
	sheet.WriteString(`</sheetData></worksheet>`)
	write("xl/worksheets/sheet1.xml", sheet.String())

	if err := zw.Close(); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func serialFor(y, m int) float64 {
	days := time.Date(y, time.Month(m), 1, 0, 0, 0, 0, time.UTC).Sub(excelEpoch).Hours() / 24
	return days
}

func TestParseM4Index(t *testing.T) {
	xlsx := buildXLSX("Broad",
		[]float64{serialFor(1967, 1), serialFor(1967, 2)},
		[]float64{100, 100.6775})

	points, err := ParseM4Index(xlsx)
	if err != nil {
		t.Fatalf("ParseM4Index: %v", err)
	}
	if len(points) != 2 {
		t.Fatalf("got %d points, want 2", len(points))
	}
	if !points[0].Date.Equal(time.Date(1967, 1, 1, 0, 0, 0, 0, time.UTC)) || points[0].Index != 100 {
		t.Errorf("unexpected first point: %v %v", points[0].Date, points[0].Index)
	}
	if points[1].Index != 100.6775 {
		t.Errorf("unexpected second index: %v", points[1].Index)
	}
}

func TestParseM4IndexMissingSheet(t *testing.T) {
	xlsx := buildXLSX("Narrow", []float64{serialFor(1967, 1)}, []float64{100})
	if _, err := ParseM4Index(xlsx); err == nil {
		t.Error("expected error for missing Broad sheet")
	}
}

func TestParseM4IndexNotAZip(t *testing.T) {
	if _, err := ParseM4Index([]byte("not a zip")); err == nil {
		t.Error("expected error for non-zip input")
	}
}

func TestFetchM4Index(t *testing.T) {
	xlsx := buildXLSX("Broad", []float64{serialFor(1967, 1)}, []float64{100})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/amfm/Divisia.xlsx" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(xlsx)
	}))
	defer srv.Close()

	points, err := New(WithBaseURL(srv.URL)).FetchM4Index()
	if err != nil {
		t.Fatalf("FetchM4Index: %v", err)
	}
	if len(points) != 1 || points[0].Index != 100 {
		t.Errorf("unexpected points: %+v", points)
	}
}

func TestFetchM4IndexBadStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()
	if _, err := New(WithBaseURL(srv.URL)).FetchM4Index(); err == nil {
		t.Error("expected error for non-200 status")
	}
}
