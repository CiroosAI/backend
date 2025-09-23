package users

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"project/database"
	"project/models"
	"project/utils"

	"gorm.io/gorm"
)

type AddBankAccountRequest struct {
	BankID        uint   `json:"bank_id"`
	AccountName   string `json:"account_name"`
	AccountNumber string `json:"account_number"`
}

func AddBankAccountHandler(w http.ResponseWriter, r *http.Request) {
	var req AddBankAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Not valid request"})
		return
	}

	req.AccountName = strings.TrimSpace(req.AccountName)
	req.AccountNumber = strings.TrimSpace(req.AccountNumber)

	if req.BankID == 0 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Bank tidak tersedia saat ini"})
		return
	}

	uid, ok := utils.GetUserID(r)
	if !ok || uid == 0 {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Unauthorized"})
		return
	}

	db := database.DB

	// Validate bank exists and Active
	var bank models.Bank
	if err := db.First(&bank, req.BankID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Bank yang dipilih tidak tersedia"})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan sistem, silakan coba lagi"})
		return
	}
	if bank.Status != "Active" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Bank yang dipilih tidak tersedia"})
		return
	}

	// Count user bank accounts (limit 3)
	var cnt int64
	if err := db.Model(&models.BankAccount{}).Where("user_id = ?", uid).Count(&cnt).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan sistem, silakan coba lagi"})
		return
	}
	if cnt >= 3 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Anda sudah mencapai batas maksimal 3 rekening bank"})
		return
	}

	// Duplicate check: user_id + bank_id + account_number
	var dup models.BankAccount
	if err := db.Where("user_id = ? AND bank_id = ? AND account_number = ?", uid, req.BankID, req.AccountNumber).First(&dup).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan sistem, silakan coba lagi"})
			return
		}
	} else {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Rekening ini sudah pernah didaftarkan"})
		return
	}

	// Validate account name: 3-100 chars, letters and spaces
	if len(req.AccountName) < 3 || len(req.AccountName) > 100 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Nama rekening harus 3-100 karakter dan hanya berisi huruf"})
		return
	}
	if ok, _ := regexp.MatchString(`^[A-Za-z ]+$`, req.AccountName); !ok {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Nama rekening harus 3-100 karakter dan hanya berisi huruf"})
		return
	}

	// Validate account number: 5-20, alphanumeric
	if len(req.AccountNumber) < 5 || len(req.AccountNumber) > 20 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Nomor rekening tidak valid"})
		return
	}
	if ok, _ := regexp.MatchString(`^[A-Za-z0-9]+$`, req.AccountNumber); !ok {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Nomor rekening tidak valid"})
		return
	}

	acc := models.BankAccount{
		UserID:        uid,
		BankID:        req.BankID,
		AccountName:   req.AccountName,
		AccountNumber: req.AccountNumber,
	}

	if err := db.Create(&acc).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan sistem, silakan coba lagi"})
		return
	}

	// Load bank for response
	_ = db.First(&bank, acc.BankID).Error

	utils.WriteJSON(w, http.StatusCreated, utils.APIResponse{
		Success: true,
		Message: "Rekening berhasil ditambahkan",
		Data: map[string]interface{}{
			"bank_account": map[string]interface{}{
				"id":             acc.ID,
				"bank_name":      bank.Name,
				"bank_code":      bank.Code,
				"account_name":   acc.AccountName,
				"account_number": acc.AccountNumber,
			},
		},
	})
}

// GET /api/users/bank or /api/users/bank/{id}
func GetBankAccountHandler(w http.ResponseWriter, r *http.Request) {
	uid, ok := utils.GetUserID(r)
	if !ok || uid == 0 {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Unauthorized"})
		return
	}
	db := database.DB
	path := r.URL.Path
	parts := strings.Split(path, "/")
	var idStr string
	if len(parts) >= 5 {
		idStr = parts[4]
	}
	if idStr == "" {
		// List all bank accounts for user
		var accounts []models.BankAccount
		if err := db.Where("user_id = ?", uid).Find(&accounts).Error; err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal mengambil data rekening"})
			return
		}
		var resp []map[string]interface{}
		for _, acc := range accounts {
			var bank models.Bank
			db.First(&bank, acc.BankID)
			resp = append(resp, map[string]interface{}{
				"id":             acc.ID,
				"account_name":   acc.AccountName,
				"account_number": acc.AccountNumber,
				"bank_id":        acc.BankID,
				"bank_name":      bank.Name,
			})
		}
		utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
			Success: true,
			Message: "Berhasil mengambil data rekening",
			Data: map[string]interface{}{
				"bank_account": resp,
			},
		})
		return
	}
	// Get by id
	var acc models.BankAccount
	if err := db.Where("user_id = ? AND id = ?", uid, idStr).First(&acc).Error; err != nil {
		utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{Success: false, Message: "Rekening tidak ditemukan"})
		return
	}
	var bank models.Bank
	db.First(&bank, acc.BankID)
	resp := []map[string]interface{}{
		{
			"id":             acc.ID,
			"account_name":   acc.AccountName,
			"account_number": acc.AccountNumber,
			"bank_id":        acc.BankID,
			"bank_name":      bank.Name,
		},
	}
	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Berhasil mengambil data rekening",
		Data: map[string]interface{}{
			"bank_account": resp,
		},
	})
}

// PUT /api/users/bank
func EditBankAccountHandler(w http.ResponseWriter, r *http.Request) {
	uid, ok := utils.GetUserID(r)
	if !ok || uid == 0 {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Unauthorized"})
		return
	}
	var req struct {
		ID            uint   `json:"id"`
		AccountName   string `json:"account_name"`
		AccountNumber string `json:"account_number"`
		BankID        uint   `json:"bank_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == 0 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Not valid request"})
		return
	}
	if req.AccountName == "" && req.AccountNumber == "" && req.BankID == 0 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Minimum one field must be filled"})
		return
	}
	db := database.DB
	var acc models.BankAccount
	if err := db.Where("user_id = ? AND id = ?", uid, req.ID).First(&acc).Error; err != nil {
		utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{Success: false, Message: "Rekening tidak ditemukan"})
		return
	}
	update := map[string]interface{}{}
	if req.AccountName != "" {
		update["account_name"] = req.AccountName
	}
	if req.AccountNumber != "" {
		update["account_number"] = req.AccountNumber
	}
	if req.BankID != 0 {
		update["bank_id"] = req.BankID
	}
	if len(update) == 0 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Minimum one field must be filled"})
		return
	}
	if err := db.Model(&acc).Updates(update).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal mengupdate rekening"})
		return
	}
	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{Success: true, Message: "Rekening berhasil diupdate"})
}

// DELETE /api/users/bank
func DeleteBankAccountHandler(w http.ResponseWriter, r *http.Request) {
	uid, ok := utils.GetUserID(r)
	if !ok || uid == 0 {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Unauthorized"})
		return
	}
	var req struct {
		ID uint `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == 0 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Not valid request"})
		return
	}
	db := database.DB
	if err := db.Where("user_id = ? AND id = ?", uid, req.ID).Delete(&models.BankAccount{}).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal menghapus rekening"})
		return
	}
	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{Success: true, Message: "Rekening berhasil dihapus"})
}
