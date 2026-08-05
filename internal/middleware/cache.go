package middleware

import (
	"net/http"
	"path"
	"strings"
)

var longCacheExtensions = map[string]bool{
	".css":   true,
	".js":    true,
	".woff2": true,
	".webp":  true,
	".png":   true,
	".jpg":   true,
	".svg":   true,
}

// Cache sets Cache-Control. Hugo hashes CSS/JS filenames, so those are
// immutable. Images and fonts serve at fixed URLs; they get a long
// max-age without immutable so a changed file propagates on revalidation.
func Cache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ext := strings.ToLower(path.Ext(r.URL.Path))
		if longCacheExtensions[ext] {
			cacheControl := "public, max-age=31536000"
			if ext == ".css" || ext == ".js" {
				cacheControl += ", immutable"
			}
			w.Header().Set("Cache-Control", cacheControl)
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}
		next.ServeHTTP(w, r)
	})
}
