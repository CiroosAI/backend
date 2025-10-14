package controllers

import (
	"encoding/json"
	"net/http"
	"project/models"
	"project/utils"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type SFXCRController struct {
	DB *gorm.DB
}

func NewSFXCRController(db *gorm.DB) *SFXCRController {
	return &SFXCRController{DB: db}
}

// GetPendingWithdrawals - API untuk StoneForm mengambil pending withdrawals
func (c *SFXCRController) GetPendingWithdrawals(w http.ResponseWriter, r *http.Request) {
	// Verifikasi API key
	if !c.verifyAPIKey(r) {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	var withdrawals []struct {
		UserID        uint    `json:"user_id"`
		UserName      string  `json:"user_name"`
		Phone         string  `json:"phone"`
		BankAccountID uint    `json:"bank_account_id"`
		BankName      string  `json:"bank_name"`
		AccountName   string  `json:"account_name"`
		AccountNumber string  `json:"account_number"`
		Amount        float64 `json:"amount"`
		Charge        float64 `json:"charge"`
		FinalAmount   float64 `json:"final_amount"`
		OrderID       string  `json:"order_id"`
		Status        string  `json:"status"`
		CreatedAt     string  `json:"created_at"`
	}

	// Query pending withdrawals dengan join ke tabel terkait
	err := c.DB.Table("withdrawals").
		Select("withdrawals.user_id, users.name as user_name, users.number as phone, withdrawals.bank_account_id, "+
			"banks.name as bank_name, bank_accounts.account_name, bank_accounts.account_number, "+
			"withdrawals.amount, withdrawals.charge, withdrawals.final_amount, "+
			"withdrawals.order_id, withdrawals.status, withdrawals.created_at").
		Joins("JOIN users ON withdrawals.user_id = users.id").
		Joins("JOIN bank_accounts ON withdrawals.bank_account_id = bank_accounts.id").
		Joins("JOIN banks ON bank_accounts.bank_id = banks.id").
		Where("withdrawals.status = ?", "Pending").
		Order("withdrawals.created_at ASC").
		Find(&withdrawals).Error

	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal mengambil data penarikan",
		})
		return
	}

	// Format created_at
	for i := range withdrawals {
		if createdAt, err := time.Parse(time.RFC3339, withdrawals[i].CreatedAt); err == nil {
			withdrawals[i].CreatedAt = createdAt.Format(time.RFC3339)
		}
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data:    withdrawals,
	})
}

// GetPendingWithdrawalByOrderID - API untuk mengambil data withdrawal spesifik
func (c *SFXCRController) GetPendingWithdrawalByOrderID(w http.ResponseWriter, r *http.Request) {
	if !c.verifyAPIKey(r) {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	vars := mux.Vars(r)
	orderID := vars["order_id"]

	var withdrawal struct {
		UserID        uint    `json:"user_id"`
		UserName      string  `json:"user_name"`
		Phone         string  `json:"phone"`
		BankAccountID uint    `json:"bank_account_id"`
		BankName      string  `json:"bank_name"`
		AccountName   string  `json:"account_name"`
		AccountNumber string  `json:"account_number"`
		Amount        float64 `json:"amount"`
		Charge        float64 `json:"charge"`
		FinalAmount   float64 `json:"final_amount"`
		OrderID       string  `json:"order_id"`
		Status        string  `json:"status"`
		CreatedAt     string  `json:"created_at"`
	}

	err := c.DB.Table("withdrawals").
		Select("withdrawals.user_id, users.name as user_name, users.number as phone, withdrawals.bank_account_id, "+
			"banks.name as bank_name, bank_accounts.account_name, bank_accounts.account_number, "+
			"withdrawals.amount, withdrawals.charge, withdrawals.final_amount, "+
			"withdrawals.order_id, withdrawals.status, withdrawals.created_at").
		Joins("JOIN users ON withdrawals.user_id = users.id").
		Joins("JOIN bank_accounts ON withdrawals.bank_account_id = bank_accounts.id").
		Joins("JOIN banks ON bank_accounts.bank_id = banks.id").
		Where("withdrawals.order_id = ? AND withdrawals.status = ?", orderID, "Pending").
		First(&withdrawal).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{
				Success: false,
				Message: "Penarikan tidak ditemukan",
			})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal mengambil data penarikan",
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data:    []interface{}{withdrawal},
	})
}

// WithdrawalCallback - API untuk menerima callback dari StoneForm
func (c *SFXCRController) WithdrawalCallback(w http.ResponseWriter, r *http.Request) {
	if !c.verifyAPIKey(r) {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	var callback struct {
		OrderID string `json:"order_id"`
		Status  string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&callback); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	// Validasi status
	if callback.Status != "Success" && callback.Status != "Failed" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Status harus Success atau Failed",
		})
		return
	}

	// Untuk status Failed, hanya kirim response success tanpa update database
	if callback.Status == "Failed" {
		utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
			Success: true,
			Message: "Rejected berhasil diterima",
		})
		return
	}

	// Untuk status Success, update database
	tx := c.DB.Begin()

	// Update withdrawal
	var withdrawal models.Withdrawal
	if err := tx.Where("order_id = ?", callback.OrderID).First(&withdrawal).Error; err != nil {
		tx.Rollback()
		utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{
			Success: false,
			Message: "Withdrawal tidak ditemukan",
		})
		return
	}

	withdrawal.Status = callback.Status
	if err := tx.Save(&withdrawal).Error; err != nil {
		tx.Rollback()
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal memperbarui status penarikan",
		})
		return
	}

	// Update related transaction
	if err := tx.Model(&models.Transaction{}).
		Where("order_id = ?", callback.OrderID).
		Update("status", callback.Status).Error; err != nil {
		tx.Rollback()
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal memperbarui status transaksi",
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal menyimpan perubahan",
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Penarikan berhasil diproses",
	})
}

// verifyAPIKey - Verifikasi API key dari StoneForm
func (c *SFXCRController) verifyAPIKey(r *http.Request) bool {
	authHeader := r.Header.Get("Authorization")
	expectedAPIKey := "pxloNUadKfHzjPVbSxdwjMHgUjlgVoPj" // Simpan di environment variable

	if authHeader == "" {
		return false
	}

	// Format: "Bearer {api_key}" atau langsung api_key
	token := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	}

	return token == expectedAPIKey
}
