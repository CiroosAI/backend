package admins

import (
	"net/http"
	"strconv"
	"time"

	"project/database"
	"project/models"
	"project/utils"
)

// GET /api/admin/user-spins
func UserSpinsHandler(w http.ResponseWriter, r *http.Request) {
	db := database.DB

	// Pagination
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	search := r.URL.Query().Get("search")
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	// Base queries
	query := db.
		Table("user_spins AS us").
		Joins("JOIN users u ON us.user_id = u.id").
		Joins("JOIN spin_prizes sp ON us.prize_id = sp.id")

	countQuery := db.
		Table("user_spins AS us").
		Joins("JOIN users u ON us.user_id = u.id").
		Joins("JOIN spin_prizes sp ON us.prize_id = sp.id")

	// Search by users.name or users.number
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("u.name LIKE ? OR u.number LIKE ?", like, like)
		countQuery = countQuery.Where("u.name LIKE ? OR u.number LIKE ?", like, like)
	}

	// Aggregates (overall)
	var totalWins int64
	if err := db.Model(&models.UserSpin{}).Count(&totalWins).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Terjadi kesalahan sistem, silakan coba lagi",
		})
		return
	}
	type paidAgg struct {
		TotalPaid float64
	}
	var agg paidAgg
	if err := db.Table("user_spins").Select("COALESCE(SUM(amount), 0) as total_paid").Scan(&agg).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Terjadi kesalahan sistem, silakan coba lagi",
		})
		return
	}

	// Row scan
	type rowScan struct {
		ID      uint
		UserID  uint
		UserName string
		Phone    string
		PrizeID uint
		Amount  float64
		Code    string
		WonAt   time.Time
	}

	var rows []rowScan
	if err := query.
		Select(`
			us.id,
			us.user_id,
			u.name AS user_name,
			u.number AS phone,
			us.prize_id,
			us.amount,
			us.code,
			us.won_at
		`).
		Order("us.won_at DESC").
		Offset(offset).
		Limit(limit).
		Scan(&rows).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Terjadi kesalahan sistem, silakan coba lagi",
		})
		return
	}

	type UserSpinResponse struct {
		ID       uint    `json:"id"`
		UserID   uint    `json:"user_id"`
		UserName string  `json:"user_name"`
		Phone    string  `json:"phone"`
		PrizeID  uint    `json:"prize_id"`
		Amount   float64 `json:"amount"`
		Code     string  `json:"code"`
		WonAt    string  `json:"won_at"`
	}

	items := make([]UserSpinResponse, 0, len(rows))
	for _, r := range rows {
		items = append(items, UserSpinResponse{
			ID:       r.ID,
			UserID:   r.UserID,
			UserName: r.UserName,
			Phone:    r.Phone,
			PrizeID:  r.PrizeID,
			Amount:   r.Amount,
			Code:     r.Code,
			WonAt:    r.WonAt.Format(time.RFC3339),
		})
	}

	// Wrap data
	type Data struct {
		TotalWins int64               `json:"total_wins"`
		TotalPaid float64             `json:"total_paid"`
		Items     []UserSpinResponse  `json:"items"`
	}
	data := Data{
		TotalWins: totalWins,
		TotalPaid: agg.TotalPaid,
		Items:     items,
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data:    data,
	})
}
