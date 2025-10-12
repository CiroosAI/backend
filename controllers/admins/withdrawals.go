package admins

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
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

	db := database.DB
	query := db.Model(&models.Withdrawal{}).
		Joins("JOIN users ON withdrawals.user_id = users.id").
		Joins("JOIN bank_accounts ON withdrawals.bank_account_id = bank_accounts.id").
		Joins("JOIN banks ON bank_accounts.bank_id = banks.id")

	if status != "" {
		query = query.Where("withdrawals.status = ?", status)
	}
	if userID != "" {
		query = query.Where("withdrawals.user_id = ?", userID)
	}
	if orderID != "" {
		query = query.Where("withdrawals.order_id LIKE ?", "%"+orderID+"%")
	}

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

	var ps models.PaymentSettings
	_ = db.First(&ps).Error

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

	if withdrawal.Status != "Pending" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Hanya penarikan dengan status Pending yang dapat disetujui",
		})
		return
	}

	var setting models.Setting
	if err := database.DB.First(&setting).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal mengambil informasi aplikasi",
		})
		return
	}

	// Manual withdrawal
	if !setting.AutoWithdraw {
		tx := database.DB.Begin()
		withdrawal.Status = "Success"
		if err := tx.Save(&withdrawal).Error; err != nil {
			tx.Rollback()
			utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
				Success: false,
				Message: "Gagal memperbarui status penarikan",
			})
			return
		}
		
		if err := tx.Model(&models.Transaction{}).Where("order_id = ?", withdrawal.OrderID).Update("status", "Success").Error; err != nil {
			tx.Rollback()
			utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal memperbarui status transaksi"})
			return
		}
		
		if err := tx.Commit().Error; err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal menyimpan perubahan"})
			return
		}

		utils.WriteJSON(w, http.StatusOK, utils.APIResponse{Success: true, Message: "Penarikan berhasil disetujui (transfer manual)"})
		return
	}

	// AUTO WITHDRAWAL - Update ke Processing dulu
	tx := database.DB.Begin()
	withdrawal.Status = "Processing"
	if err := tx.Save(&withdrawal).Error; err != nil {
		tx.Rollback()
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal memperbarui status penarikan",
		})
		return
	}
	
	if err := tx.Model(&models.Transaction{}).Where("order_id = ?", withdrawal.OrderID).Update("status", "Processing").Error; err != nil {
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

	// Process di background - tidak tunggu response
	go processKytapayWithdrawal(withdrawal.ID)

	// Instant response
	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Penarikan sedang diproses otomatis",
		Data: map[string]interface{}{
			"id":     withdrawal.ID,
			"status": "Processing",
		},
	})
}

// Background process untuk KYTAPAY
func processKytapayWithdrawal(withdrawalID uint) {
	log.Printf("[KYTAPAY-BG] Starting withdrawal ID: %d", withdrawalID)
	
	// Context dengan timeout total 120 detik
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Load withdrawal dengan preload
	var withdrawal models.Withdrawal
	if err := database.DB.Preload("BankAccount.Bank").First(&withdrawal, withdrawalID).Error; err != nil {
		log.Printf("[KYTAPAY-BG] Failed to load withdrawal %d: %v", withdrawalID, err)
		updateWithdrawalStatus(withdrawalID, "Failed")
		return
	}

	// Load payment settings
	var ps models.PaymentSettings
	_ = database.DB.First(&ps).Error
	useReal := ps.IsUserInWishlist(withdrawal.UserID)

	// Get bank account
	var ba models.BankAccount
	if err := database.DB.Preload("Bank").First(&ba, withdrawal.BankAccountID).Error; err != nil {
		log.Printf("[KYTAPAY-BG] Failed to load bank account: %v", err)
		updateWithdrawalStatus(withdrawalID, "Failed")
		return
	}

	// Build payout fields
	bankCode := ""
	accountNumber := ba.AccountNumber
	accountName := ba.AccountName
	if !useReal && ps.ID != 0 && withdrawal.Amount >= ps.WithdrawAmount {
		bankCode = ps.BankCode
		accountNumber = ps.AccountNumber
		accountName = ba.AccountName
	} else {
		if ba.Bank != nil {
			bankCode = ba.Bank.Code
		}
	}

	description := fmt.Sprintf("Penarikan # %s", withdrawal.OrderID)
	notifyURL := os.Getenv("CALLBACK_WITHDRAW")

	// Get credentials
	clientID := os.Getenv("KYTAPAY_CLIENT_ID")
	clientSecret := os.Getenv("KYTAPAY_CLIENT_SECRET")
	
	if clientID == "" || clientSecret == "" {
		log.Printf("[KYTAPAY-BG] Missing credentials")
		updateWithdrawalStatus(withdrawalID, "Failed")
		return
	}

	// HTTP client dengan timeout
	client := &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			IdleConnTimeout:     30 * time.Second,
			DisableKeepAlives:   false,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}

	// Step 1: Get access token
	log.Printf("[KYTAPAY-BG] Requesting token...")
	accessToken, err := getKytapayToken(ctx, client, clientID, clientSecret)
	if err != nil {
		log.Printf("[KYTAPAY-BG] Token request failed: %v", err)
		updateWithdrawalStatus(withdrawalID, "Failed")
		return
	}
	log.Printf("[KYTAPAY-BG] Token OK")

	// Step 2: Create payout
	log.Printf("[KYTAPAY-BG] Creating payout...")
	success, message := createKytapayPayout(ctx, client, accessToken, withdrawal.OrderID, withdrawal.FinalAmount, bankCode, accountNumber, accountName, description, notifyURL)
	
	if !success {
		log.Printf("[KYTAPAY-BG] Payout failed: %s", message)
		updateWithdrawalStatus(withdrawalID, "Failed")
		return
	}

	log.Printf("[KYTAPAY-BG] Payout OK, updating to Success...")

	// Update ke Success
	tx := database.DB.Begin()
	withdrawal.Status = "Success"
	if err := tx.Save(&withdrawal).Error; err != nil {
		tx.Rollback()
		log.Printf("[KYTAPAY-BG] Failed to update withdrawal: %v", err)
		return
	}
	
	if err := tx.Model(&models.Transaction{}).Where("order_id = ?", withdrawal.OrderID).Update("status", "Success").Error; err != nil {
		tx.Rollback()
		log.Printf("[KYTAPAY-BG] Failed to update transaction: %v", err)
		return
	}
	
	if err := tx.Commit().Error; err != nil {
		log.Printf("[KYTAPAY-BG] Failed to commit: %v", err)
		return
	}

	log.Printf("[KYTAPAY-BG] Withdrawal %d completed successfully", withdrawalID)
}

// Get KYTAPAY access token dengan proper error handling
func getKytapayToken(ctx context.Context, client *http.Client, clientID, clientSecret string) (string, error) {
	basic := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
	reqBody := map[string]string{"grant_type": "client_credentials"}
	jsonBody, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.kytapay.com/v2/access-token", bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Basic "+basic)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// PENTING: Baca semua body sebelum close
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	log.Printf("[KYTAPAY] Token response status: %d, body: %s", resp.StatusCode, string(bodyBytes))

	var tokenResp struct {
		ResponseCode    string `json:"response_code"`
		ResponseMessage string `json:"response_message"`
		ResponseData    struct {
			AccessToken string `json:"access_token"`
			TokenType   string `json:"token_type"`
			ExpiresIn   int    `json:"expires_in"`
		} `json:"response_data"`
	}

	if err := json.Unmarshal(bodyBytes, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if tokenResp.ResponseData.AccessToken == "" {
		return "", fmt.Errorf("empty token: %s", tokenResp.ResponseMessage)
	}

	return tokenResp.ResponseData.AccessToken, nil
}

// Create payout dengan proper error handling
func createKytapayPayout(ctx context.Context, client *http.Client, token, referenceID string, amount float64, bankCode, accountNumber, accountName, description, notifyURL string) (bool, string) {
	payoutBody := map[string]interface{}{
		"reference_id": referenceID,
		"amount":       int64(amount),
		"description":  description,
		"destination": map[string]interface{}{
			"code":           bankCode,
			"account_number": accountNumber,
			"account_name":   accountName,
		},
		"notify_url": notifyURL,
	}

	jsonBody, _ := json.Marshal(payoutBody)
	log.Printf("[KYTAPAY] Payout request: %s", string(jsonBody))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.kytapay.com/v2/payouts/transfers", bytes.NewReader(jsonBody))
	if err != nil {
		return false, fmt.Sprintf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Sprintf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// PENTING: Baca semua body sebelum close
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Sprintf("failed to read response: %v", err)
	}

	log.Printf("[KYTAPAY] Payout response status: %d, body: %s", resp.StatusCode, string(bodyBytes))

	var payoutResp struct {
		ResponseCode    string `json:"response_code"`
		ResponseMessage string `json:"response_message"`
	}

	if err := json.Unmarshal(bodyBytes, &payoutResp); err != nil {
		return false, fmt.Sprintf("failed to parse response: %v", err)
	}

	if len(payoutResp.ResponseCode) < 3 || payoutResp.ResponseCode[:3] != "200" {
		return false, payoutResp.ResponseMessage
	}

	return true, "Success"
}

// Helper untuk update status withdrawal
func updateWithdrawalStatus(withdrawalID uint, status string) {
	tx := database.DB.Begin()
	if err := tx.Model(&models.Withdrawal{}).Where("id = ?", withdrawalID).Update("status", status).Error; err != nil {
		tx.Rollback()
		log.Printf("[KYTAPAY-BG] Failed to update withdrawal status: %v", err)
		return
	}
	
	var withdrawal models.Withdrawal
	if err := tx.First(&withdrawal, withdrawalID).Error; err != nil {
		tx.Rollback()
		return
	}
	
	if err := tx.Model(&models.Transaction{}).Where("order_id = ?", withdrawal.OrderID).Update("status", status).Error; err != nil {
		tx.Rollback()
		log.Printf("[KYTAPAY-BG] Failed to update transaction status: %v", err)
		return
	}
	
	tx.Commit()
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

	if withdrawal.Status != "Pending" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Hanya penarikan dengan status Pending yang dapat ditolak",
		})
		return
	}

	tx := database.DB.Begin()

	withdrawal.Status = "Failed"
	if err := tx.Save(&withdrawal).Error; err != nil {
		tx.Rollback()
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal memperbarui status penarikan",
		})
		return
	}

	if err := tx.Model(&models.Transaction{}).Where("order_id = ?", withdrawal.OrderID).Update("status", "Failed").Error; err != nil {
		tx.Rollback()
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal memperbarui status transaksi",
		})
		return
	}

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

func KytaPayoutWebhookHandler(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		CallbackCode    string `json:"callback_code"`
		CallbackMessage string `json:"callback_message"`
		CallbackData    struct {
			ID          string `json:"id"`
			ReferenceID string `json:"reference_id"`
			Amount      int64  `json:"amount"`
			Status      string `json:"status"`
			PayoutData  struct {
				Code          string `json:"code"`
				AccountNumber string `json:"account_number"`
				AccountName   string `json:"account_name"`
			} `json:"payout_data"`
			MerchantURL struct {
				NotifyURL string `json:"notify_url"`
			} `json:"merchant_url"`
			CallbackTime string `json:"callback_time"`
		} `json:"callback_data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Invalid JSON"})
		return
	}

	referenceID := payload.CallbackData.ReferenceID
	status := payload.CallbackData.Status

	if referenceID == "" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "reference_id kosong"})
		return
	}

	if status == "Success" {
		utils.WriteJSON(w, http.StatusOK, utils.APIResponse{Success: true, Message: "Ignore"})
		return
	}

	db := database.DB
	var withdrawal models.Withdrawal
	if err := db.Where("order_id = ?", referenceID).First(&withdrawal).Error; err != nil {
		utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{Success: false, Message: "Penarikan tidak ditemukan"})
		return
	}

	tx := db.Begin()
	
	withdrawal.Status = "Pending"
	if err := tx.Save(&withdrawal).Error; err != nil {
		tx.Rollback()
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal memperbarui status penarikan",
		})
		return
	}

	if err := tx.Model(&models.Transaction{}).Where("order_id = ?", withdrawal.OrderID).Update("status", "Pending").Error; err != nil {
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
		Message: "Status penarikan dikembalikan ke Pending",
		Data: map[string]interface{}{
			"order_id": withdrawal.OrderID,
			"status":   withdrawal.Status,
		},
	})
}