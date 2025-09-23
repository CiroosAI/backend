package admins

import (
	"net/http"
	"strconv"
	"time"

	"project/database"
	"project/models"
	"project/utils"
)

type PaymentResponse struct {
	ID             uint   `json:"id"`
	InvestmentID   uint   `json:"investment_id"`
	ReferenceID    string `json:"reference_id"`
	OrderID        string `json:"order_id"`
	PaymentMethod  string `json:"payment_method"`
	PaymentChannel string `json:"payment_channel"`
	PaymentCode    string `json:"payment_code"`
	Status         string `json:"status"`
	ExpiredAt      string `json:"expired_at"`
	CreatedAt      string `json:"created_at"`
}

func GetPayments(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	investmentId := r.URL.Query().Get("investmentId")
	userId := r.URL.Query().Get("userId")
	status := r.URL.Query().Get("status")
	startDate := r.URL.Query().Get("startDate")
	endDate := r.URL.Query().Get("endDate")

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Start query
	db := database.DB
	query := db.Model(&models.Payment{})

	// Apply filters
	if investmentId != "" {
		query = query.Where("investment_id = ?", investmentId)
	}
	if userId != "" {
		// Join with investments table to filter by user_id
		query = query.Joins("JOIN investments ON payments.investment_id = investments.id").
			Where("investments.user_id = ?", userId)
	}
	if status != "" {
		query = query.Where("payments.status = ?", status)
	}

	// Apply date filters if provided
	jakartaLoc, _ := time.LoadLocation("Asia/Jakarta")
	if startDate != "" {
		startTime, err := time.ParseInLocation("2006-01-02", startDate, jakartaLoc)
		if err == nil {
			query = query.Where("payments.created_at >= ?", startTime)
		}
	}
	if endDate != "" {
		endTime, err := time.ParseInLocation("2006-01-02", endDate, jakartaLoc)
		if err == nil {
			// Add one day to get to the start of the next day in Jakarta time
			endTime = endTime.AddDate(0, 0, 1)
			query = query.Where("payments.created_at < ?", endTime)
		}
	}

	var payments []models.Payment
	query.Offset(offset).
		Limit(limit).
		Order("payments.created_at DESC").
		Find(&payments)

	// Transform to response format
	var response []PaymentResponse
	for _, p := range payments {
		response = append(response, PaymentResponse{
			ID:             p.ID,
			InvestmentID:   p.InvestmentID,
			ReferenceID:    utils.GetStringValue(p.ReferenceID),
			OrderID:        p.OrderID,
			PaymentMethod:  utils.GetStringValue(p.PaymentMethod),
			PaymentChannel: utils.GetStringValue(p.PaymentChannel),
			PaymentCode:    utils.GetStringValue(p.PaymentCode),
			Status:         p.Status,
			ExpiredAt:      p.ExpiredAt.Format(time.RFC3339),
			CreatedAt:      p.CreatedAt.Format(time.RFC3339),
		})
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data:    response,
	})
}
