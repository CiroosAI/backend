package users

import (
	"net/http"
	"project/database"
	"project/models"
	"project/utils"
	"strconv"
	"strings"
	"time"
)

// use apiResponse from info.go

// use writeJSON from info.go

// GET /api/users/transaction/{type}
func GetTransactionHistory(w http.ResponseWriter, r *http.Request) {
	uid, ok := utils.GetUserID(r)
	if !ok || uid == 0 {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Unauthorized"})
		return
	}

	// Get type from query param ?type= or fallback to URL path segment
	txType := strings.TrimSpace(r.URL.Query().Get("type"))
	if txType == "" {
		path := r.URL.Path
		parts := strings.Split(path, "/")
		if len(parts) >= 5 {
			txType = strings.TrimSpace(parts[4])
		}
	}

	// Pagination: page + limit
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	page := 1
	limit := 25
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
		limit = l
	}
	offset := (page - 1) * limit

	db := database.DB
	var transactions []models.Transaction
	query := db.Where("user_id = ?", uid)
	if txType != "" && txType != "null" {
		query = query.Where("transaction_type = ?", txType)
	}
	if err := query.Order("id DESC").Limit(limit).Offset(offset).Find(&transactions).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Database error"})
		return
	}

	// Map transactions to DTO including created_at
	type transactionDTO struct {
		ID              uint    `json:"id"`
		UserID          uint    `json:"user_id"`
		Amount          float64 `json:"amount"`
		Charge          float64 `json:"charge"`
		OrderID         string  `json:"order_id"`
		TransactionFlow string  `json:"transaction_flow"`
		TransactionType string  `json:"transaction_type"`
		Message         *string `json:"message,omitempty"`
		Status          string  `json:"status"`
		CreatedAt       string  `json:"created_at"`
	}

	items := make([]transactionDTO, 0, len(transactions))
	for _, t := range transactions {
		items = append(items, transactionDTO{
			ID:              t.ID,
			UserID:          t.UserID,
			Amount:          t.Amount,
			Charge:          t.Charge,
			OrderID:         t.OrderID,
			TransactionFlow: t.TransactionFlow,
			TransactionType: t.TransactionType,
			Message:         t.Message,
			Status:          t.Status,
			CreatedAt:       t.CreatedAt.Format(time.RFC3339),
		})
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data:    map[string]interface{}{"transactions": items},
	})
}
