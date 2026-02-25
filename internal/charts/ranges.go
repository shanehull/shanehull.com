package charts

import "time"

// CalculateRangeStart converts a UI range parameter to a start date.
// Returns nil for "max" to fetch all available data.
// Used by chart tools to determine observation_start for FRED queries.
func CalculateRangeStart(rangeParam string) *time.Time {
	now := time.Now()
	switch rangeParam {
	case "1y":
		t := now.AddDate(-1, 0, 0)
		return &t
	case "5y":
		t := now.AddDate(-5, 0, 0)
		return &t
	case "10y":
		t := now.AddDate(-10, 0, 0)
		return &t
	case "20y":
		t := now.AddDate(-20, 0, 0)
		return &t
	case "50y":
		t := now.AddDate(-50, 0, 0)
		return &t
	default: // "max" - no limit
		return nil
	}
}
