package admins

import (
	"encoding/json"
	"net/http"
	"strconv"

	"project/database"
	"project/models"
	"project/utils"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type SpinPrizeResponse struct {
	ID           uint    `json:"id"`
	Amount       float64 `json:"amount"`
	Code         string  `json:"code"`
	ChanceWeight int     `json:"chance_weight"`
	Chance       float64 `json:"chance"`
	Status       string  `json:"status"`
	TotalWins    int64   `json:"total_wins"`
	TotalPaid    float64 `json:"total_paid"`
}

type UpdateSpinPrizeRequest struct {
	Amount       float64 `json:"amount"`
	Code         string  `json:"code"`
	ChanceWeight int     `json:"chance_weight"`
	Status       string  `json:"status"`
}

func calculateChances(prizes []models.SpinPrize) []SpinPrizeResponse {
	// Calculate total weight
	totalWeight := 0
	for _, p := range prizes {
		totalWeight += p.ChanceWeight
	}

	// Calculate chances and create response
	var response []SpinPrizeResponse
	for _, prize := range prizes {
		chance := float64(0)
		if totalWeight > 0 {
			chance = float64(prize.ChanceWeight) / float64(totalWeight) * 100
		}
		response = append(response, SpinPrizeResponse{
			ID:           prize.ID,
			Amount:       prize.Amount,
			Code:         prize.Code,
			ChanceWeight: prize.ChanceWeight,
			Chance:       utils.RoundFloat(chance, 2), // Round to 2 decimal places
			Status:       prize.Status,
		})
	}
	return response
}

func GetSpinPrizes(w http.ResponseWriter, r *http.Request) {
	var prizes []models.SpinPrize
	if err := database.DB.Find(&prizes).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal mengambil data hadiah spin",
		})
		return
	}

	// Base response with chances
	response := calculateChances(prizes)

	// Build aggregates: total wins and total paid per prize
	type prizeAgg struct {
		PrizeID uint
		Wins    int64
		Paid    float64
	}
	var aggs []prizeAgg
	if err := database.DB.
		Table("user_spins").
		Select("prize_id, COUNT(*) as wins, COALESCE(SUM(amount), 0) as paid").
		Group("prize_id").
		Scan(&aggs).Error; err == nil {
		aggMap := make(map[uint]prizeAgg, len(aggs))
		for _, a := range aggs {
			aggMap[a.PrizeID] = a
		}
		for i := range response {
			if a, ok := aggMap[response[i].ID]; ok {
				response[i].TotalWins = a.Wins
				response[i].TotalPaid = a.Paid
			}
		}
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data:    response,
	})
}

func UpdateSpinPrize(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Invalid prize ID",
		})
		return
	}

	var req UpdateSpinPrizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	var prize models.SpinPrize
	if err := database.DB.First(&prize, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{
				Success: false,
				Message: "Hadiah tidak ditemukan",
			})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal mengambil data hadiah",
		})
		return
	}

	// Update prize details
	if err := database.DB.Model(&prize).Updates(map[string]interface{}{
		"amount":        req.Amount,
		"code":          req.Code,
		"chance_weight": req.ChanceWeight,
		"status":        req.Status,
	}).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal memperbarui hadiah",
		})
		return
	}

	// Get all prizes to calculate new chances
	var allPrizes []models.SpinPrize
	if err := database.DB.Find(&allPrizes).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal mengambil data hadiah",
		})
		return
	}

	// Calculate new chances
	response := calculateChances(allPrizes)

	// Find the updated prize in response
	var updatedPrize SpinPrizeResponse
	for _, p := range response {
		if p.ID == prize.ID {
			updatedPrize = p
			break
		}
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Hadiah Spin berhasil diperbarui",
		Data:    updatedPrize,
	})
}
