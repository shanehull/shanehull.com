package abs

import (
	"testing"
	"time"
)

const aweFixture = `{
  "data": {
    "structures": [{
      "dimensions": {
        "observation": [
          {"id":"TIME_PERIOD","values":[{"id":"2024-S1"},{"id":"2024-S2"}]}
        ]
      }
    }],
    "dataSets": [{
      "series": {
        "0:0": {"observations": {"0":[1600.0],"1":[1650.0]}}
      }
    }]
  }
}`

func TestParseSDMX(t *testing.T) {
	observations, err := parseSDMX([]byte(aweFixture))
	if err != nil {
		t.Fatal(err)
	}
	if len(observations) != 2 || observations["2024-S2"] != 1650 {
		t.Fatalf("unexpected observations: %+v", observations)
	}
	points, err := toPoints(observations)
	if err != nil {
		t.Fatal(err)
	}
	if len(points) != 2 || points[1].Value != 1650 {
		t.Fatalf("unexpected points: %+v", points)
	}
}

func TestParsePeriod(t *testing.T) {
	cases := map[string]time.Time{
		"2026-Q2": time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		"1994-S2": time.Date(1994, 11, 1, 0, 0, 0, 0, time.UTC),
		"1994-S1": time.Date(1994, 5, 1, 0, 0, 0, 0, time.UTC),
	}
	for in, want := range cases {
		got, err := parsePeriod(in)
		if err != nil {
			t.Fatalf("parsePeriod(%q): %v", in, err)
		}
		if !got.Equal(want) {
			t.Fatalf("parsePeriod(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestCellDate(t *testing.T) {
	// Excel serial 32387 is 1988-09-01; ISO strings are also accepted.
	for in, want := range map[string]string{
		"32387":               "1988-09-01",
		"1988-09-01 00:00:00": "1988-09-01",
		"1988-09-01":          "1988-09-01",
	} {
		got, ok := cellDate(in)
		if !ok {
			t.Fatalf("cellDate(%q) not ok", in)
		}
		if got.Format("2006-01-02") != want {
			t.Fatalf("cellDate(%q) = %s, want %s", in, got.Format("2006-01-02"), want)
		}
	}
	if _, ok := cellDate("nonsense"); ok {
		t.Fatal("expected not ok for nonsense")
	}
}

func TestFindColumn(t *testing.T) {
	header := []string{"Date", "Residential land and dwellings;", "Other"}
	if got := findColumn(header, "Residential land and dwellings"); got != 1 {
		t.Fatalf("findColumn = %d, want 1", got)
	}
	if got := findColumn(header, "Missing"); got != -1 {
		t.Fatalf("findColumn missing = %d, want -1", got)
	}
}
