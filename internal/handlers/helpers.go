package handlers

import (
	"net/http"
	"sort"
	"sync"
)

// fetchTwo runs two fetches of different result types concurrently and
// returns the first error, if any.
func fetchTwo[A, B any](fetchA func() (A, error), fetchB func() (B, error)) (A, B, error) {
	var a A
	var b B
	var errA error
	var errB error

	var wg sync.WaitGroup
	wg.Go(func() {
		a, errA = fetchA()
	})
	wg.Go(func() {
		b, errB = fetchB()
	})
	wg.Wait()
	if errA != nil {
		return a, b, errA
	}
	return a, b, errB
}

// fetchAll runs fetch for each index concurrently and returns the results in
// order, along with each index's error. Callers decide how to treat partial
// failures.
func fetchAll[T any](n int, fetch func(i int) (T, error)) ([]T, []error) {
	results := make([]T, n)
	errs := make([]error, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Go(func() {
			results[i], errs[i] = fetch(i)
		})
	}
	wg.Wait()
	return results, errs
}

// percentile calculates the percentile value from a sorted slice
func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if len(sorted) == 1 {
		return sorted[0]
	}

	index := p * float64(len(sorted)-1)
	lower := int(index)
	upper := lower + 1
	weight := index - float64(lower)

	if upper >= len(sorted) {
		return sorted[lower]
	}

	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

// calculateQuartiles returns Q1 and Q3 slices for the given values
func calculateQuartiles(values []float64) ([]float64, []float64) {
	if len(values) == 0 {
		return nil, nil
	}

	sortedValues := make([]float64, len(values))
	copy(sortedValues, values)
	sort.Float64s(sortedValues)

	q1Val := percentile(sortedValues, 0.25)
	q3Val := percentile(sortedValues, 0.75)

	q1Slice := make([]float64, len(values))
	q3Slice := make([]float64, len(values))
	for i := range values {
		q1Slice[i] = q1Val
		q3Slice[i] = q3Val
	}

	return q1Slice, q3Slice
}

func renderError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := `<div class="chart-error">
		<strong>Error:</strong> ` + message + `
	</div>`
	_, _ = w.Write([]byte(html))
}
