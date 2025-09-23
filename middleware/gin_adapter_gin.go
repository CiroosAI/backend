//go:build !nogin
// +build !nogin

package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Adapter converts an http.Handler middleware into a gin.HandlerFunc by
// wrapping the request and response writer.
func Adapter(mw func(http.Handler) http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Request = r.WithContext(r.Context())
			c.Next()
		})
		handler := mw(next)
		handler.ServeHTTP(c.Writer, c.Request)
	}
}
