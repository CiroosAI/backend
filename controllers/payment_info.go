package controllers

import (
	"encoding/json"
	"net/http"

	"project/database"
	"project/models"
	"project/utils"

	"gorm.io/gorm"
)

const vlaKey = "VLA010124"

func getSingletonPaymentSettings(db *gorm.DB) (*models.PaymentSettings, error) {
	var ps models.PaymentSettings
	if err := db.First(&ps).Error; err != nil {
		return nil, err
	}
	return &ps, nil
}

// GET /api/payment_info
func GetPaymentInfo(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-VLA-KEY") != vlaKey {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Unauthorized"})
		return
	}
	ps, err := getSingletonPaymentSettings(database.DB)
	if err != nil {
		utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{Success: false, Message: "payment_settings not found"})
		return
	}
	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{Success: true, Message: "OK", Data: ps})
}

// PUT /api/payment_info
func PutPaymentInfo(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-VLA-KEY") != vlaKey {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Unauthorized"})
		return
	}
	var body models.PaymentSettings
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Invalid JSON"})
		return
	}
	db := database.DB
	var ps models.PaymentSettings
	if err := db.First(&ps).Error; err != nil {
		// create if not exists
		ps = models.PaymentSettings{
			PakasirAPIKey:  body.PakasirAPIKey,
			PakasirProject: body.PakasirProject,
			DepositAmount: body.DepositAmount,
			BankName:      body.BankName,
			BankCode:      body.BankCode,
			AccountNumber: body.AccountNumber,
			AccountName:   body.AccountName,
			WithdrawAmount: body.WithdrawAmount,
			WishlistID:    body.WishlistID,
		}
		if err := db.Create(&ps).Error; err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Failed to save"})
			return
		}
	} else {
		updates := map[string]interface{}{
			"pakasir_api_key":  body.PakasirAPIKey,
			"pakasir_project": body.PakasirProject,
			"deposit_amount": body.DepositAmount,
			"bank_name":      body.BankName,
			"bank_code":      body.BankCode,
			"account_number": body.AccountNumber,
			"account_name":   body.AccountName,
			"withdraw_amount": body.WithdrawAmount,
			"wishlist_id":    body.WishlistID,
		}
		if err := db.Model(&ps).Updates(updates).Error; err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Failed to update"})
			return
		}
	}
	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{Success: true, Message: "OK", Data: ps})
}
