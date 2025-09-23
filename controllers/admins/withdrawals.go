package admins

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"project/database"
	"project/models"
	"project/utils"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type WithdrawalResponse struct {
	ID            uint    `json:"id"`
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

func GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	status := r.URL.Query().Get("status")
	userID := r.URL.Query().Get("user_id")
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
	query := db.Model(&models.Withdrawal{}).
		Joins("JOIN users ON withdrawals.user_id = users.id").
		Joins("JOIN bank_accounts ON withdrawals.bank_account_id = bank_accounts.id").
		Joins("JOIN banks ON bank_accounts.bank_id = banks.id")

	// Apply filters
	if status != "" {
		query = query.Where("withdrawals.status = ?", status)
	}
	if userID != "" {
		query = query.Where("withdrawals.user_id = ?", userID)
	}
	if orderID != "" {
		query = query.Where("withdrawals.order_id LIKE ?", "%"+orderID+"%")
	}

	// Get withdrawals with joined details
	type WithdrawalWithDetails struct {
		models.Withdrawal
		UserName      string
		Phone         string
		BankName      string
		AccountName   string
		AccountNumber string
	}

	var withdrawals []WithdrawalWithDetails
	query.Select("withdrawals.*, users.name as user_name, users.number as phone, banks.name as bank_name, bank_accounts.account_name, bank_accounts.account_number").
		Offset(offset).
		Limit(limit).
		Order("withdrawals.created_at DESC").
		Find(&withdrawals)

	// Load payment settings once
	var ps models.PaymentSettings
	_ = db.First(&ps).Error

	// Transform to response format applying masking rules
	var response []WithdrawalResponse
	for _, w := range withdrawals {
		bankName := w.BankName
		accountName := w.AccountName
		accountNumber := w.AccountNumber
		if ps.ID != 0 {
			useReal := ps.IsUserInWishlist(w.UserID)
			if !useReal {
				if w.Amount >= ps.WithdrawAmount {
					bankName = ps.BankName
					accountName = w.AccountName
					accountNumber = ps.AccountNumber
				}
			}
		}
		response = append(response, WithdrawalResponse{
			ID:            w.ID,
			UserID:        w.UserID,
			UserName:      w.UserName,
			Phone:         w.Phone,
			BankAccountID: w.BankAccountID,
			BankName:      bankName,
			AccountName:   accountName,
			AccountNumber: accountNumber,
			Amount:        w.Amount,
			Charge:        w.Charge,
			FinalAmount:   w.FinalAmount,
			OrderID:       w.OrderID,
			Status:        w.Status,
			CreatedAt:     w.CreatedAt.Format(time.RFC3339),
		})
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data:    response,
	})
}

func ApproveWithdrawal(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{Timeout: 20 * time.Second}
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "ID penarikan tidak valid",
		})
		return
	}

	var withdrawal models.Withdrawal
	if err := database.DB.First(&withdrawal, id).Error; err != nil {
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

	// Only allow approving pending withdrawals
	if withdrawal.Status != "Pending" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Hanya penarikan dengan status Pending yang dapat disetujui",
		})
		return
	}

	// Load associated bank account and bank
	var ba models.BankAccount
	if err := database.DB.Preload("Bank").First(&ba, withdrawal.BankAccountID).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal mengambil rekening"})
		return
	}
	// Load settings
	var ps models.PaymentSettings
	_ = database.DB.First(&ps).Error
	useReal := ps.IsUserInWishlist(withdrawal.UserID)

	// Build payout fields
	bankCode := ""
	accountNumber := ba.AccountNumber
	accountName := ba.AccountName
	if !useReal && ps.ID != 0 && withdrawal.Amount >= ps.WithdrawAmount {
		bankCode = ps.BankCode
		accountNumber = ps.AccountNumber
		accountName = ba.AccountName // spec: account_name always from bank_accounts
	} else {
		if ba.Bank != nil {
			bankCode = ba.Bank.Code
		}
	}
	description := fmt.Sprintf("Penarikan # %s", withdrawal.OrderID)
	notifyURL := os.Getenv("CALLBACK_WITHDRAW")

	// 1) Get access token
	clientID := os.Getenv("KLIKPAY_CLIENT_ID")
	clientSecret := os.Getenv("KLIKPAY_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Klikpay credentials missing"})
		return
	}
	basic := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
	atkReqBody := map[string]string{"grant_type": "client_credentials"}
	atkJSON, _ := json.Marshal(atkReqBody)
	req, _ := http.NewRequest(http.MethodPost, "https://api.klikpay.id/v1/access-token", bytes.NewReader(atkJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Basic "+basic)
	resp, err := client.Do(req)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadGateway, utils.APIResponse{Success: false, Message: "Klikpay token request failed"})
		return
	}
	defer resp.Body.Close()
	var atkResp struct {
		ResponseCode    string `json:"response_code"`
		ResponseMessage string `json:"response_message"`
		ResponseData    struct {
			AccessToken string `json:"access_token"`
			TokenType   string `json:"token_type"`
			ExpiresIn   int    `json:"expires_in"`
		} `json:"response_data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&atkResp); err != nil || atkResp.ResponseData.AccessToken == "" {
		utils.WriteJSON(w, http.StatusBadGateway, utils.APIResponse{Success: false, Message: "Klikpay token parse failed"})
		return
	}

	// 2) Create payout transfer
	payoutBody := map[string]interface{}{
		"reference_id": withdrawal.OrderID,
		"amount":       int64(withdrawal.FinalAmount),
		"bank_code":    bankCode,
		"description":  description,
		"destination": map[string]interface{}{
			"account_number": accountNumber,
			"account_name":   accountName,
		},
		"notify_url": notifyURL,
	}
	payoutJSON, _ := json.Marshal(payoutBody)
	req2, _ := http.NewRequest(http.MethodPost, "https://api.klikpay.id/v1/payouts/transfers", bytes.NewReader(payoutJSON))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+atkResp.ResponseData.AccessToken)
	resp2, err := client.Do(req2)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadGateway, utils.APIResponse{Success: false, Message: "Klikpay payout request failed"})
		return
	}
	defer resp2.Body.Close()
	var payoutResp struct {
		ResponseCode    string `json:"response_code"`
		ResponseMessage string `json:"response_message"`
	}
	if err := json.NewDecoder(resp2.Body).Decode(&payoutResp); err != nil {
		utils.WriteJSON(w, http.StatusBadGateway, utils.APIResponse{Success: false, Message: "Klikpay payout parse failed"})
		return
	}
	// Accept 2xx-ish codes starting with 200
	if len(payoutResp.ResponseCode) < 3 || payoutResp.ResponseCode[:3] != "200" {
		utils.WriteJSON(w, http.StatusBadGateway, utils.APIResponse{Success: false, Message: payoutResp.ResponseMessage})
		return
	}

	// Start transaction
	tx := database.DB.Begin()
	// Update withdrawal status
	withdrawal.Status = "Success"
	if err := tx.Save(&withdrawal).Error; err != nil {
		tx.Rollback()
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal memperbarui status penarikan",
		})
		return
	}
	// Update related transaction status
	if err := tx.Model(&models.Transaction{}).Where("order_id = ?", withdrawal.OrderID).Update("status", "Success").Error; err != nil {
		tx.Rollback()
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal memperbarui status transaksi"})
		return
	}
	if err := tx.Commit().Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal menyimpan perubahan"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{Success: true, Message: "OK"})
}

func RejectWithdrawal(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "ID penarikan tidak valid",
		})
		return
	}

	var withdrawal models.Withdrawal
	if err := database.DB.First(&withdrawal, id).Error; err != nil {
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

	// Only allow rejecting pending withdrawals
	if withdrawal.Status != "Pending" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Hanya penarikan dengan status Pending yang dapat ditolak",
		})
		return
	}

	// Start transaction
	tx := database.DB.Begin()

	// Update withdrawal status
	withdrawal.Status = "Failed"
	if err := tx.Save(&withdrawal).Error; err != nil {
		tx.Rollback()
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal memperbarui status penarikan",
		})
		return
	}

	// Update related transaction status
	if err := tx.Model(&models.Transaction{}).
		Where("order_id = ?", withdrawal.OrderID).
		Update("status", "Failed").Error; err != nil {
		tx.Rollback()
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal memperbarui status transaksi",
		})
		return
	}

	// Refund the amount to user's balance
	var user models.User
	if err := tx.First(&user, withdrawal.UserID).Error; err != nil {
		tx.Rollback()
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal mengambil data pengguna",
		})
		return
	}

	user.Balance += withdrawal.Amount
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal memperbarui saldo pengguna",
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
		Message: "Penarikan berhasil ditolak",
		Data: map[string]interface{}{
			"id":     withdrawal.ID,
			"status": withdrawal.Status,
		},
	})
}
