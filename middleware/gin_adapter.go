//go:build !nogin
// +build !nogin

package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GinAdapter converts an http.Handler middleware into a gin.HandlerFunc by
// wrapping the request and response writer.
func GinAdapter(mw func(http.Handler) http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a handler that continues the gin chain
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// copy the gin request context into the request and proceed with gin
			c.Request = r.WithContext(r.Context())
			c.Next()
		})
		handler := mw(next)
		handler.ServeHTTP(c.Writer, c.Request)
	}
}
