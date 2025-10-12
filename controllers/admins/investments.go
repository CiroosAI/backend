package admins

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"project/database"
	"project/models"
	"project/utils"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type InvestmentResponse struct {
	ID            uint    `json:"id"`
	UserID        uint    `json:"user_id"`
	UserName      string  `json:"username"`
	Phone         string  `json:"phone"`
	ProductID     uint    `json:"product_id"`
	ProductName   string  `json:"product_name"`
	CategoryID    uint    `json:"category_id"`
	CategoryName  string  `json:"category_name"`
	Amount        float64 `json:"amount"`
	Duration      int     `json:"duration"`
	DailyProfit   float64 `json:"daily_profit"`
	TotalPaid     int     `json:"total_paid"`
	TotalReturned float64 `json:"total_returned"`
	LastReturnAt  string  `json:"last_return_at,omitempty"`
	NextReturnAt  string  `json:"next_return_at,omitempty"`
	OrderID       string  `json:"order_id"`
	Status        string  `json:"status"`
	CreatedAt     string  `json:"created_at"`
}

func GetInvestments(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	productID := r.URL.Query().Get("product_id")
	status := r.URL.Query().Get("status")
	orderID := r.URL.Query().Get("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Start query
	db := database.DB
	query := db.Model(&models.Investment{}).
		Joins("JOIN products ON investments.product_id = products.id").
		Joins("JOIN categories ON investments.category_id = categories.id")

	// Apply filters
	if productID != "" {
		query = query.Where("investments.product_id = ?", productID)
	}
	if status != "" {
		query = query.Where("investments.status = ?", status)
	}
	if orderID != "" {
		query = query.Where("investments.order_id LIKE ?", "%"+orderID+"%")
	}

	// Get investments with product and category details
	type InvestmentWithProduct struct {
		models.Investment
		ProductName  string
		CategoryName string
	}

	var investments []InvestmentWithProduct
	query.Select("investments.*, products.name as product_name, categories.name as category_name").
		Offset(offset).
		Limit(limit).
		Order("investments.created_at DESC").
		Find(&investments)

	// Prepare user IDs to fetch names in batch
	userIDsSet := make(map[uint]struct{})
	for _, inv := range investments {
		userIDsSet[inv.UserID] = struct{}{}
	}
	var userIDs []uint
	for id := range userIDsSet {
		userIDs = append(userIDs, id)
	}

	// Fetch users and build a map[id]user
	usersByID := make(map[uint]models.User, len(userIDs))
	if len(userIDs) > 0 {
		var users []models.User
		db.Select("id, name, number").Where("id IN ?", userIDs).Find(&users)
		for _, u := range users {
			usersByID[u.ID] = u
		}
	}

	// Transform to response format
	var response []InvestmentResponse
	for _, inv := range investments {
		response = append(response, InvestmentResponse{
			ID:            inv.ID,
			UserID:        inv.UserID,
			UserName:      usersByID[inv.UserID].Name,
			Phone:         usersByID[inv.UserID].Number,
			ProductID:     inv.ProductID,
			ProductName:   inv.ProductName,
			CategoryID:    inv.CategoryID,
			CategoryName:  inv.CategoryName,
			Amount:        inv.Amount,
			Duration:      inv.Duration,
			DailyProfit:   inv.DailyProfit,
			TotalPaid:     inv.TotalPaid,
			TotalReturned: inv.TotalReturned,
			LastReturnAt:  formatTimePtr(inv.LastReturnAt),
			NextReturnAt:  formatTimePtr(inv.NextReturnAt),
			OrderID:       inv.OrderID,
			Status:        inv.Status,
			CreatedAt:     inv.CreatedAt.Format(time.RFC3339),
		})
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data:    response,
	})
}

func GetInvestmentDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Invalid investment ID",
		})
		return
	}

	// Get investment with product and category details
	type InvestmentWithProduct struct {
		models.Investment
		ProductName  string
		CategoryName string
	}

	var investment InvestmentWithProduct
	err = database.DB.Model(&models.Investment{}).
		Joins("JOIN products ON investments.product_id = products.id").
		Joins("JOIN categories ON investments.category_id = categories.id").
		Select("investments.*, products.name as product_name, categories.name as category_name").
		Where("investments.id = ?", id).
		First(&investment).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{
				Success: false,
				Message: "Investment not found",
			})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Terjadi kesalahan sistem, silakan coba lagi",
		})
		return
	}

	// Fetch user name
	var user models.User
	_ = database.DB.Select("id, name, number").First(&user, investment.UserID).Error

	response := InvestmentResponse{
		ID:            investment.ID,
		UserID:        investment.UserID,
		UserName:      user.Name,
		Phone:         user.Number,
		ProductID:     investment.ProductID,
		ProductName:   investment.ProductName,
		CategoryID:    investment.CategoryID,
		CategoryName:  investment.CategoryName,
		Amount:        investment.Amount,
		Duration:      investment.Duration,
		DailyProfit:   investment.DailyProfit,
		TotalPaid:     investment.TotalPaid,
		TotalReturned: investment.TotalReturned,
		LastReturnAt:  formatTimePtr(investment.LastReturnAt),
		NextReturnAt:  formatTimePtr(investment.NextReturnAt),
		OrderID:       investment.OrderID,
		Status:        investment.Status,
		CreatedAt:     investment.CreatedAt.Format(time.RFC3339),
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data:    response,
	})
}

type UpdateInvestmentStatusRequest struct {
	Status string `json:"status"`
}

func UpdateInvestmentStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "ID investasi tidak valid",
		})
		return
	}

	var req UpdateInvestmentStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Format data tidak valid",
		})
		return
	}

	// Validate status
	validStatuses := map[string]bool{
		"Suspended": true,
		"Cancelled": true,
		"Completed": true,
		"Running":   true,
	}

	if !validStatuses[req.Status] {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Status tidak valid",
		})
		return
	}

	var investment models.Investment
	if err := database.DB.First(&investment, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{
				Success: false,
				Message: "Investasi tidak ditemukan",
			})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal mengambil data investasi",
		})
		return
	}

	// If changing from Pending to Running, set next_return_at
	if investment.Status == "Pending" || investment.Status == "Cancelled" && req.Status == "Running" {
		nextReturn := time.Now().Add(24 * time.Hour)
		investment.NextReturnAt = &nextReturn
	}

	// Update status
	investment.Status = req.Status

	if err := database.DB.Save(&investment).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal memperbarui status investasi",
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Status investasi berhasil diperbarui",
		Data: map[string]interface{}{
			"id":     investment.ID,
			"status": investment.Status,
		},
	})
}

func formatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}
