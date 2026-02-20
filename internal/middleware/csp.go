package middleware

import (
	"net/http"
)

func CSP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CSP header
		csp := "script-src 'self' https://cdn.jsdelivr.net/npm/htmx.org@2.0.7/ https://static.cloudflareinsights.com/ https://cdn.jsdelivr.net/npm/chart.js https://cdn.jsdelivr.net/npm/chartjs-plugin-zoom@2.1.0/dist/chartjs-plugin-zoom.min.js; object-src 'none'; base-uri 'self';"
		w.Header().Set("Content-Security-Policy", csp)

		next.ServeHTTP(w, r)
	})
}
