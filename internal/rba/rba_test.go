package rba

import (
	"testing"
)

const d2Fixture = `D2 LENDING AND CREDIT AGGREGATES
Title,Credit; Owner-occupier housing,Credit; Investor housing
Description,Owner-occupier housing credit,Investor housing credit
Frequency,Monthly,Monthly
Type,Original,Original
Units,$ billion,$ billion

Source,RBA,RBA
Publication date,30-Oct-2026,30-Oct-2026
Series ID,DLCACOHN,DLCACIHN
31/01/1990,10.0,5.0
28/02/1990,11.0,6.0
`

func TestParseColumn(t *testing.T) {
	points, err := parseColumn([]byte(d2Fixture), "Credit; Owner-occupier housing")
	if err != nil {
		t.Fatal(err)
	}
	if len(points) != 2 || points[0].Value != 10 || points[1].Value != 11 {
		t.Fatalf("unexpected points: %+v", points)
	}
	if got := points[0].Date.Format("2006-01-02"); got != "1990-01-31" {
		t.Fatalf("want 1990-01-31, got %s", got)
	}
	if _, err := parseColumn([]byte(d2Fixture), "No such column"); err == nil {
		t.Fatal("expected an error for a missing column")
	}
}

func TestSumSeries(t *testing.T) {
	owner, err := parseColumn([]byte(d2Fixture), "Credit; Owner-occupier housing")
	if err != nil {
		t.Fatal(err)
	}
	investor, err := parseColumn([]byte(d2Fixture), "Credit; Investor housing")
	if err != nil {
		t.Fatal(err)
	}
	got := sumSeries(owner, investor)
	if len(got) != 2 {
		t.Fatalf("want 2 points, got %d", len(got))
	}
	if got[0].Value != 15 || got[1].Value != 17 {
		t.Fatalf("unexpected sums: %+v", got)
	}
}
