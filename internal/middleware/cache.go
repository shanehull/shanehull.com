package middleware

import (
	"net/http"
	"path"
	"strings"
)

var fingerprintedExtensions = map[string]bool{
	".css": true,
	".js":  true,
}

// Cache sets Cache-Control. Hugo hashes CSS/JS filenames so they are
// immutable; images and fonts serve at fixed URLs and must revalidate.
func Cache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ext := strings.ToLower(path.Ext(r.URL.Path))
		if fingerprintedExtensions[ext] {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}
		next.ServeHTTP(w, r)
	})
}
