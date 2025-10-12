package routes

import (
	"encoding/json"
	"net/http"
	"time"

	"project/controllers"
	"project/controllers/users"
	"project/controllers/admins"
	"project/middleware"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func optionsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func InitRouter() *mux.Router {
	r := mux.NewRouter()

	// Add CORS middleware
	r.Use(func(next http.Handler) http.Handler {
		return handlers.CORS(
			handlers.AllowedOrigins([]string{"https://stoneform.site", "http://localhost:3000"}),
			handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
			handlers.AllowedHeaders([]string{"Content-Type", "Authorization", "X-VLA-KEY", "X-CRON-KEY"}),
			handlers.AllowCredentials(),
		)(next)
	})

	api := r.PathPrefix("/api").Subrouter()

	// Add catch-all OPTIONS handler for CORS preflight
	api.PathPrefix("/").HandlerFunc(optionsHandler).Methods(http.MethodOptions)

	// Rate limiter untuk cron: 1000/jam
	cronLimiter := middleware.NewIPRateLimiter(1000, time.Hour)
	// Rate limiter untuk webhook: 500/ip, whitelist, sliding window
	webhookLimiter := middleware.NewWebhookLimiter(500, time.Hour, []string{"127.0.0.1" /* tambahkan IP whitelist di sini */})

	// Cron endpoint for daily returns (protected via X-CRON-KEY header)
	api.Handle("/cron/daily-returns", cronLimiter.Middleware(http.HandlerFunc(users.CronDailyReturnsHandler))).Methods(http.MethodPost)

	// Kytapay webhook (no auth, whitelist, sliding window)
	api.Handle("/payments/kyta/webhook", webhookLimiter.Middleware(http.HandlerFunc(users.KytaWebhookHandler))).Methods(http.MethodPost)

	api.Handle("/payouts/kyta/webhook", webhookLimiter.Middleware(http.HandlerFunc(admins.KytaPayoutWebhookHandler))).Methods(http.MethodPost)

	// Example protected endpoint using JWT middleware
	api.Handle("/ping", middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "pong",
		})
	}))).Methods(http.MethodGet)

	// Public application info
	api.Handle("/info", http.HandlerFunc(controllers.InfoPublicHandler)).Methods(http.MethodGet)

	// Health check endpoint for Docker health checks
	api.Handle("/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
			"service":   "stoneform-api",
		})
	})).Methods(http.MethodGet)

	// Payment settings endpoints (protected by static header)
	api.Handle("/payment_info", http.HandlerFunc(controllers.GetPaymentInfo)).Methods(http.MethodGet)
	api.Handle("/payment_info", http.HandlerFunc(controllers.PutPaymentInfo)).Methods(http.MethodPut)

	// Delegasi semua route users ke file users.go
	UsersRoutes(api)

	// Setup admin routes
	SetAdminRoutes(api)

	return r
}
