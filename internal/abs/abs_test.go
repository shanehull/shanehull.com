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
