package middleware

import (
	"net/http"
	"path"
	"strings"
)

var immutableExtensions = map[string]bool{
	".css":   true,
	".js":    true,
	".woff2": true,
	".webp":  true,
	".png":   true,
	".jpg":   true,
	".svg":   true,
}

// Cache sets Cache-Control headers for static assets. Fingerprinted
// and immutable asset files get a long max-age; HTML is never cached.
func Cache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ext := strings.ToLower(path.Ext(r.URL.Path))
		if immutableExtensions[ext] {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}
		next.ServeHTTP(w, r)
	})
}
