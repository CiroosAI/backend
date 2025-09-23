package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"project/utils"
)

// generateRequestID creates a short random request id
func generateRequestID() string {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return hex.EncodeToString(b)
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// SecurityHeadersMiddleware sets CORS and security headers. Behavior is env-driven.
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	// Configurable values
	env := strings.ToLower(getenv("ENV", "development"))
	allowedOrigins := getenv("CORS_ALLOWED_ORIGINS", "*")
	hsts := getenv("SEC_HSTS", "false")
	csp := getenv("SEC_CSP", "default-src 'none'; frame-ancestors 'none'; base-uri 'self';")

	// Build a list for CORS matches
	origins := strings.Split(allowedOrigins, ",")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORS
		origin := r.Header.Get("Origin")
		allowed := false
		if allowedOrigins == "*" {
			allowed = true
		} else if origin != "" {
			for _, o := range origins {
				if strings.TrimSpace(o) == origin {
					allowed = true
					break
				}
			}
		}
		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Requested-With, X-Request-ID")
		}

		// Security headers
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		if env != "development" {
			w.Header().Set("Content-Security-Policy", csp)
		}
		if hsts == "true" {
			// 1 year HSTS
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		// Preflight handling
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequestIDMiddleware injects a request id into context and response headers
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get("X-Request-ID")
		if rid == "" {
			rid = generateRequestID()
		}
		w.Header().Set("X-Request-ID", rid)
		ctx := context.WithValue(r.Context(), utils.RequestIDKey, rid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// TimeoutMiddleware cancels the request context after a configured timeout
func TimeoutMiddleware(next http.Handler) http.Handler {
	timeoutSec := atoi(getenv("REQ_TIMEOUT_SEC", "10"))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Duration(timeoutSec)*time.Second)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RecoveryMiddleware recovers from panics, logs securely and returns generic 500
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				// Secure logging: include request id and minimal context
				ridVal := r.Context().Value(utils.RequestIDKey)
				rid := ""
				if s, ok := ridVal.(string); ok {
					rid = s
				}
				// Log the panic with request context (do not expose internals to client)
				log.Printf("PANIC recovered: request_id=%s method=%s path=%s panic=%v", rid, r.Method, r.URL.Path, rec)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Internal server error", "request_id": rid})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// Simple in-memory metrics and suspicious activity tracker
var (
	metricsMu sync.Mutex
	// track last N response times per route
	routeTimes = make(map[string][]time.Duration)
	// track recent suspicious activity per ip
	suspiciousMu sync.Mutex
	suspicious   = make(map[string]int)
)

// MetricsMiddleware measures response time and tracks slow responses
func MetricsMiddleware(next http.Handler) http.Handler {
	slowThresholdMs := atoi(getenv("METRIC_SLOW_MS", "800"))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		elapsed := time.Since(start)
		// record route times
		metricsMu.Lock()
		key := r.Method + " " + r.URL.Path
		arr := routeTimes[key]
		if len(arr) >= 100 {
			arr = arr[1:]
		}
		arr = append(arr, elapsed)
		routeTimes[key] = arr
		metricsMu.Unlock()

		if elapsed > time.Duration(slowThresholdMs)*time.Millisecond {
			// increment suspicious counter for the IP
			ip := r.RemoteAddr
			suspiciousMu.Lock()
			suspicious[ip] = suspicious[ip] + 1
			suspiciousMu.Unlock()
		}
	})
}

// SuspiciousActivityMiddleware flags IPs with repeated slow responses or other signals
func SuspiciousActivityMiddleware(next http.Handler) http.Handler {
	threshold := atoi(getenv("SUSPICIOUS_THRESHOLD", "10"))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		suspiciousMu.Lock()
		count := suspicious[ip]
		suspiciousMu.Unlock()
		if count >= threshold {
			// Return a generic 429 to slow down potential enumeration
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Too many requests"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Helper: atoi with default
func atoi(s string) int {
	v, _ := strconv.Atoi(s)
	if v <= 0 {
		return 0
	}
	return v
}
