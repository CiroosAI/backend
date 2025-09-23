package middleware

import (
	"net/http"
	"os"
	"strconv"
)

// MaxBodyMiddleware enforces a maximum request body size read from env var MAX_BODY_BYTES (in bytes)
// default is 1<<20 (1 MiB)
func MaxBodyMiddleware(next http.Handler) http.Handler {
	max := int64(1 << 20) // 1 MiB default
	if s := os.Getenv("MAX_BODY_BYTES"); s != "" {
		if v, err := strconv.ParseInt(s, 10, 64); err == nil && v > 0 {
			max = v
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// apply MaxBytesReader to limit request body size
		r.Body = http.MaxBytesReader(w, r.Body, max)
		// call handler
		next.ServeHTTP(w, r)
	})
}
