package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"project/utils"
)

func writeJSON(w http.ResponseWriter, status int, resp map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authz := r.Header.Get("Authorization")
		if authz == "" || !strings.HasPrefix(authz, "Bearer ") {
			writeJSON(w, http.StatusUnauthorized, map[string]interface{}{
				"success": false,
				"message": "Unauthorized",
			})
			return
		}
		tokenStr := strings.TrimSpace(strings.TrimPrefix(authz, "Bearer "))
		// Use shared validation which checks signature and registered claims
		token, claims, err := utils.ValidateAccessToken(tokenStr)
		if err != nil {
			if strings.Contains(err.Error(), "expired") {
				writeJSON(w, http.StatusUnauthorized, map[string]interface{}{
					"success": false,
					"message": "Sesi anda telah habis, silahkan login kembali.",
				})
				return
			}
			writeJSON(w, http.StatusUnauthorized, map[string]interface{}{
				"success": false,
				"message": "Invalid token",
			})
			return
		}

		_ = token // token kept for potential future use

		// Extract user ID
		var userID uint
		if rawID, ok := claims["id"]; ok {
			switch v := rawID.(type) {
			case float64:
				userID = uint(v)
			case int:
				userID = uint(v)
			case string:
				var n uint
				_, _ = fmt.Sscanf(v, "%d", &n)
				userID = n
			}
		}

		// Extract role
		var role string
		if rStr, ok := claims["role"].(string); ok {
			role = rStr
		}

		// block admin role from user endpoints (keep existing behavior)
		if role == "admin" {
			writeJSON(w, http.StatusForbidden, map[string]interface{}{
				"success": false,
				"message": "Access denied",
			})
			return
		}

		ctx := context.WithValue(r.Context(), utils.UserIDKey, userID)
		ctx = context.WithValue(ctx, utils.UserRoleKey, role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
