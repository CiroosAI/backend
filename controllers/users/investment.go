package users

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"project/database"
	"project/models"
	"project/utils"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type KytaAccessTokenResponse struct {
	ResponseCode    string `json:"response_code"`
	ResponseMessage string `json:"response_message"`
	ResponseData    struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
		RequestTime string `json:"request_time"`
	} `json:"response_data"`
}

type KytaPaymentResponse struct {
	ResponseCode    string `json:"response_code"`
	ResponseMessage string `json:"response_message"`
	ResponseData    struct {
		ID          string `json:"id"`
		ReferenceID string `json:"reference_id"`
		Amount      int64  `json:"amount"`
		PaymentData struct {
			QRString      string `json:"qr_string,omitempty"`
			BankCode      string `json:"bank_code,omitempty"`
			AccountNumber string `json:"account_number,omitempty"`
			AccountName   string `json:"account_name,omitempty"`
		} `json:"payment_data"`
		MerchantURL struct {
			NotifyURL  string `json:"notify_url"`
			SuccessURL string `json:"success_url"`
			FailedURL  string `json:"failed_url"`
		} `json:"merchant_url"`
		CheckoutURL string `json:"checkout_url"`
		ExpiresAt   string `json:"expires_at"`
		RequestTime string `json:"request_time"`
	} `json:"response_data"`
}

type CreateInvestmentRequest struct {
	ProductID      uint    `json:"product_id"`
	Amount         float64 `json:"amount"`
	PaymentMethod  string  `json:"payment_method"`
	PaymentChannel string  `json:"payment_channel"`
}

// GET /api/users/investment/active
func GetActiveInvestmentsHandler(w http.ResponseWriter, r *http.Request) {
	uid, ok := utils.GetUserID(r)
	if !ok || uid == 0 {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Unauthorized"})
		return
	}
	db := database.DB
	var products []models.Product
	if err := db.Where("status = ?", "Active").Order("id ASC").Find(&products).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal mengambil produk"})
		return
	}
	var investments []models.Investment
	if err := db.Where("user_id = ? AND status IN ?", uid, []string{"Running", "Completed", "Suspended"}).Order("product_id ASC, id DESC").Find(&investments).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal mengambil investasi"})
		return
	}
	invMap := make(map[uint][]models.Investment)
	for _, inv := range investments {
		invMap[inv.ProductID] = append(invMap[inv.ProductID], inv)
	}
	resp := make(map[string]interface{})
	for _, prod := range products {
		var result []map[string]interface{}
		if invs, ok := invMap[prod.ID]; ok && len(invs) > 0 {
			for _, inv := range invs {
				m := map[string]interface{}{
					"id":             inv.ID,
					"user_id":        inv.UserID,
					"product_id":     inv.ProductID,
					"amount":         int64(inv.Amount),
					"percentage":     inv.Percentage,
					"duration":       inv.Duration,
					"daily_profit":   int64(inv.DailyProfit),
					"total_paid":     inv.TotalPaid,
					"total_returned": int64(inv.TotalReturned),
					"last_return_at": inv.LastReturnAt,
					"next_return_at": inv.NextReturnAt,
					"order_id":       inv.OrderID,
					"status":         inv.Status,
				}
				result = append(result, m)
			}
			resp[prod.Name] = result
		} else {
			resp[prod.Name] = []map[string]interface{}{}
		}
	}
	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{Success: true, Message: "Successfully", Data: resp})
}

// POST /api/users/investments
func CreateInvestmentHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateInvestmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Not valid JSON"})
		return
	}

	uid, ok := utils.GetUserID(r)
	if !ok || uid == 0 {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Unauthorized"})
		return
	}

	method := strings.ToUpper(strings.TrimSpace(req.PaymentMethod))
	channel := strings.ToUpper(strings.TrimSpace(req.PaymentChannel))
	if method != "QRIS" && method != "BANK" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Silahkan pilih metode pembayaran"})
		return
	}
	if method == "BANK" {
		allowed := map[string]struct{}{"BCA": {}, "BRI": {}, "BNI": {}, "MANDIRI": {}, "PERMATA": {}, "BNC": {}}
		if _, ok := allowed[channel]; !ok {
			utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Bank tidak valid"})
			return
		}
	}

	db := database.DB
	var product models.Product
	if err := db.Where("id = ? AND status = 'Active'", req.ProductID).First(&product).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Produk tidak ditemukan"})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan, coba lagi"})
		return
	}
	if req.Amount < product.Minimum || req.Amount > product.Maximum {
		msg := fmt.Sprintf("Nominal investasi untuk %s harus antara Rp%.0f - Rp%.0f", product.Name, product.Minimum, product.Maximum)
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: msg})
		return
	}

	if req.ProductID == 2 || req.ProductID == 3 {
		var user models.User
		if err := db.Select("total_invest").Where("id = ?", uid).First(&user).Error; err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan, coba lagi"})
			return
		}
		if user.TotalInvest < 200000 {
			msg := fmt.Sprintf("Anda harus berinvestasi pada Bintang 1 senilai Rp200.000 terlebih dahulu sebelum berinvestasi pada %s", product.Name)
			utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: msg})
			return
		}
	}

	kytapayBase := os.Getenv("KYTAPAY_BASE_URL")
	if kytapayBase == "" {
		kytapayBase = "https://api.kytapay.com/v2"
	}
	kytapayClientID := os.Getenv("KYTAPAY_CLIENT_ID")
	kytapayClientSecret := os.Getenv("KYTAPAY_CLIENT_SECRET")
	notifyURL := os.Getenv("NOTIFY_URL")
	successURL := os.Getenv("SUCCESS_URL")
	failedURL := os.Getenv("FAILED_URL")

	if kytapayClientID == "" || kytapayClientSecret == "" {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Server error"})
		return
	}

	httpClient := &http.Client{Timeout: 15 * time.Second}
	orderID := utils.GenerateOrderID(uid)
	referenceID := orderID

	accessToken, err := getKytaAccessToken(r.Context(), httpClient, kytapayBase, kytapayClientID, kytapayClientSecret)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadGateway, utils.APIResponse{Success: false, Message: "Gagal mendapatkan token akses"})
		return
	}

	var payResp *KytaPaymentResponse
	if method == "QRIS" {
		payResp, err = createKytaQRIS(r.Context(), httpClient, kytapayBase, accessToken, referenceID, int64(req.Amount), notifyURL, successURL, failedURL)
	} else {
		payResp, err = createKytaVA(r.Context(), httpClient, kytapayBase, accessToken, referenceID, int64(req.Amount), channel, notifyURL, successURL, failedURL)
	}

	if err != nil {
		utils.WriteJSON(w, http.StatusBadGateway, utils.APIResponse{Success: false, Message: "Gagal memanggil layanan pembayaran"})
		return
	}
	if payResp == nil {
		utils.WriteJSON(w, http.StatusBadGateway, utils.APIResponse{Success: false, Message: "Gagal mendapatkan jawaban dari layanan pembayaran"})
		return
	}

	daily := round3((req.Amount*product.Percentage/100.0 + req.Amount) / float64(product.Duration))

	inv := models.Investment{
		UserID:        uid,
		ProductID:     product.ID,
		Amount:        req.Amount,
		Percentage:    product.Percentage,
		Duration:      product.Duration,
		DailyProfit:   daily,
		TotalPaid:     0,
		TotalReturned: 0,
		OrderID:       orderID,
		Status:        "Pending",
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&inv).Error; err != nil {
			return err
		}

		var paymentCode *string
		var expiredAt *time.Time

		methodToSave := strings.ToUpper(method)

		if method == "QRIS" {
			if qr := strings.TrimSpace(payResp.ResponseData.PaymentData.QRString); qr != "" {
				paymentCode = &qr
			}
		} else {
			if accNum := strings.TrimSpace(payResp.ResponseData.PaymentData.AccountNumber); accNum != "" {
				paymentCode = &accNum
			}
		}

		if expiredStr := strings.TrimSpace(payResp.ResponseData.ExpiresAt); expiredStr != "" {
			if t, err := parseTimeFlexible(expiredStr); err == nil {
				tt := t.UTC()
				expiredAt = &tt
			} else {
				t := time.Now().Add(15 * time.Minute)
				expiredAt = &t
			}
		} else {
			t := time.Now().Add(15 * time.Minute)
			expiredAt = &t
		}

		payment := models.Payment{
			InvestmentID: inv.ID,
			ReferenceID: func() *string {
				x := referenceID
				return &x
			}(),
			OrderID:       inv.OrderID,
			PaymentMethod: &methodToSave,
			PaymentChannel: func() *string {
				if methodToSave == "BANK" {
					return &channel
				}
				return nil
			}(),
			PaymentCode: paymentCode,
			PaymentLink: func() *string {
				if url := strings.TrimSpace(payResp.ResponseData.CheckoutURL); url != "" {
					return &url
				}
				return nil
			}(),
			Status:    "Pending",
			ExpiredAt: expiredAt,
		}

		if err := tx.Create(&payment).Error; err != nil {
			return err
		}

		if paymentCode != nil && *paymentCode != "" {
			if err := tx.Model(&models.Payment{}).Where("id = ?", payment.ID).Update("payment_code", *paymentCode).Error; err != nil {
				return err
			}
		}

		msg := fmt.Sprintf("Investasi %s", product.Name)
		trx := models.Transaction{
			UserID:          uid,
			Amount:          inv.Amount,
			Charge:          0,
			OrderID:         inv.OrderID,
			TransactionFlow: "credit",
			TransactionType: "investment",
			Message:         &msg,
			Status:          "Pending",
		}
		if err := tx.Create(&trx).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal membuat investasi"})
		return
	}

	resp := map[string]interface{}{
		"order_id":     inv.OrderID,
		"amount":       inv.Amount,
		"product":      product.Name,
		"percentage":   product.Percentage,
		"duration":     product.Duration,
		"daily_profit": daily,
		"status":       inv.Status,
	}
	utils.WriteJSON(w, http.StatusCreated, utils.APIResponse{Success: true, Message: "Pembelian berhasil, silakan lakukan pembayaran", Data: resp})
}

// GET /api/users/investments
func ListInvestmentsHandler(w http.ResponseWriter, r *http.Request) {
	uid, ok := utils.GetUserID(r)
	if !ok || uid == 0 {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Unauthorized"})
		return
	}
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(limitStr)
	if limit < 1 || limit > 50 {
		limit = 25
	}
	offset := (page - 1) * limit

	db := database.DB
	var rows []models.Investment
	if err := db.Where("user_id = ?", uid).Order("id DESC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan"})
		return
	}
	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{Success: true, Message: "Successfully", Data: map[string]interface{}{"investments": rows}})
}

// GET /api/users/investments/{id}
func GetInvestmentHandler(w http.ResponseWriter, r *http.Request) {
	uid, ok := utils.GetUserID(r)
	if !ok || uid == 0 {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Unauthorized"})
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	var idStr string
	if len(parts) >= 4 {
		idStr = parts[3]
	}
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "ID tidak valid"})
		return
	}
	db := database.DB
	var row models.Investment
	if err := db.Where("id = ? AND user_id = ?", uint(id64), uid).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{Success: false, Message: "Data tidak ditemukan"})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan"})
		return
	}
	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{Success: true, Message: "Successfully", Data: row})
}

// GET /api/users/payment/{order_id}
func GetPaymentDetailsHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	var orderID string
	if len(parts) >= 3 {
		orderID = parts[len(parts)-1]
	}

	db := database.DB
	var payment models.Payment
	if err := db.Where("order_id = ?", orderID).First(&payment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{Success: false, Message: "Data pembayaran tidak ditemukan"})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan"})
		return
	}

	var inv models.Investment
	if err := db.Where("id = ?", payment.InvestmentID).First(&inv).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan mengambil data investasi"})
		return
	}
	var product models.Product
	if err := db.Select("name").Where("id = ?", inv.ProductID).First(&product).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan mengambil data produk"})
		return
	}
	resp := map[string]interface{}{
		"product":  product.Name,
		"order_id": payment.OrderID,
		"amount":   inv.Amount,
		"payment_code": func() interface{} {
			if payment.PaymentCode == nil {
				return nil
			}
			return *payment.PaymentCode
		}(),
		"payment_channel": func() interface{} {
			if payment.PaymentChannel == nil {
				return nil
			}
			return *payment.PaymentChannel
		}(),
		"payment_method": func() interface{} {
			if payment.PaymentMethod == nil {
				return nil
			}
			return *payment.PaymentMethod
		}(),
		"expired_at": func() interface{} {
			if payment.ExpiredAt == nil {
				return nil
			}
			return payment.ExpiredAt.UTC().Format(time.RFC3339)
		}(),
		"status": payment.Status,
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{Success: true, Message: "Successfully", Data: resp})
}

// POST /api/payments/kyta/webhook
func KytaWebhookHandler(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		CallbackCode    string `json:"callback_code"`
		CallbackMessage string `json:"callback_message"`
		CallbackData    struct {
			ID          string `json:"id"`
			ReferenceID string `json:"reference_id"`
			Amount      string  `json:"amount"`
			Status      string `json:"status"`
			PaymentType string `json:"payment_type"`
			PaymentData struct {
				QRString      string `json:"qr_string,omitempty"`
				BankCode      string `json:"bank_code,omitempty"`
				AccountNumber string `json:"account_number,omitempty"`
				AccountName   string `json:"account_name,omitempty"`
			} `json:"payment_data"`
			MerchantURL struct {
				NotifyURL  string `json:"notify_url"`
				SuccessURL string `json:"success_url"`
				FailedURL  string `json:"failed_url"`
			} `json:"merchant_url"`
			CallbackTime string `json:"callback_time"`
		} `json:"callback_data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Invalid JSON"})
		return
	}

	referenceID := strings.TrimSpace(payload.CallbackData.ReferenceID)
	status := strings.ToUpper(strings.TrimSpace(payload.CallbackData.Status))
	paymentID := strings.TrimSpace(payload.CallbackData.ID)

	if referenceID == "" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "reference_id kosong"})
		return
	}

	success := status == "SUCCESS" || status == "PAID" || status == "COMPLETED"

	db := database.DB

	var payment models.Payment
	if err := db.Where("order_id = ?", referenceID).First(&payment).Error; err != nil {
		utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{Success: false, Message: "Pembayaran tidak ditemukan"})
		return
	}

	paymentUpdates := map[string]interface{}{}
	if paymentID != "" {
		paymentUpdates["reference_id"] = paymentID
	}
	if success {
		paymentUpdates["status"] = "Success"
	} else {
		paymentUpdates["status"] = "Failed"
	}
	if len(paymentUpdates) > 0 {
		_ = db.Model(&payment).Updates(paymentUpdates).Error
	}

	var inv models.Investment
	if err := db.Where("id = ?", payment.InvestmentID).First(&inv).Error; err != nil {
		utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{Success: false, Message: "Investasi tidak ditemukan"})
		return
	}

	if inv.Status != "Pending" {
		utils.WriteJSON(w, http.StatusOK, utils.APIResponse{Success: true, Message: "Ignored"})
		return
	}

	if success {
		now := time.Now()
		next := now.Add(24 * time.Hour)
		_ = db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&models.Transaction{}).Where("order_id = ?", inv.OrderID).Updates(map[string]interface{}{"status": "Success"}).Error; err != nil {
				return err
			}
			updates := map[string]interface{}{"status": "Running", "last_return_at": nil, "next_return_at": next}
			if err := tx.Model(&inv).Updates(updates).Error; err != nil {
				return err
			}
			if err := tx.Model(&models.User{}).Where("id = ?", inv.UserID).Updates(map[string]interface{}{
				"total_invest":      gorm.Expr("total_invest + ?", inv.Amount),
				"investment_status": "Active",
			}).Error; err != nil {
				return err
			}

			var userLevel uint
			if err := tx.Model(&models.User{}).Select("level").Where("id = ?", inv.UserID).Scan(&userLevel).Error; err != nil {
				return err
			}
			if inv.ProductID > userLevel {
				if err := tx.Model(&models.User{}).Where("id = ?", inv.UserID).Update("level", inv.ProductID).Error; err != nil {
					return err
				}
			}

			var user models.User
			if err := tx.Select("id, reff_by").Where("id = ?", inv.UserID).First(&user).Error; err == nil && user.ReffBy != nil {
				var level1 models.User
				if err := tx.Select("id, reff_by, spin_ticket").Where("id = ?", *user.ReffBy).First(&level1).Error; err == nil {
					if inv.Amount >= 100000 {
						if level1.SpinTicket == nil {
							one := uint(1)
							tx.Model(&models.User{}).Where("id = ?", level1.ID).Update("spin_ticket", one)
						} else {
							tx.Model(&models.User{}).Where("id = ?", level1.ID).UpdateColumn("spin_ticket", gorm.Expr("spin_ticket + 1"))
						}
					}

					bonus1 := round3(inv.Amount * 0.10)
					tx.Model(&models.User{}).Where("id = ?", level1.ID).UpdateColumn("balance", gorm.Expr("balance + ?", bonus1))
					msg1 := "Bonus rekomendasi investor level 1"
					trx1 := models.Transaction{
						UserID:          level1.ID,
						Amount:          bonus1,
						Charge:          0,
						OrderID:         utils.GenerateOrderID(level1.ID),
						TransactionFlow: "debit",
						TransactionType: "team",
						Message:         &msg1,
						Status:          "Success",
					}
					tx.Create(&trx1)

					if level1.ReffBy != nil {
						var level2 models.User
						if err := tx.Select("id, reff_by").Where("id = ?", *level1.ReffBy).First(&level2).Error; err == nil {
							bonus2 := round3(inv.Amount * 0.05)
							tx.Model(&models.User{}).Where("id = ?", level2.ID).UpdateColumn("balance", gorm.Expr("balance + ?", bonus2))
							msg2 := "Bonus rekomendasi investor level 2"
							trx2 := models.Transaction{
								UserID:          level2.ID,
								Amount:          bonus2,
								Charge:          0,
								OrderID:         utils.GenerateOrderID(level2.ID),
								TransactionFlow: "debit",
								TransactionType: "bonus",
								Message:         &msg2,
								Status:          "Success",
							}
							tx.Create(&trx2)

							if level2.ReffBy != nil {
								var level3 models.User
								if err := tx.Select("id").Where("id = ?", *level2.ReffBy).First(&level3).Error; err == nil {
									bonus3 := round3(inv.Amount * 0.01)
									tx.Model(&models.User{}).Where("id = ?", level3.ID).UpdateColumn("balance", gorm.Expr("balance + ?", bonus3))
									msg3 := "Bonus rekomendasi investor level 3"
									trx3 := models.Transaction{
										UserID:          level3.ID,
										Amount:          bonus3,
										Charge:          0,
										OrderID:         utils.GenerateOrderID(level3.ID),
										TransactionFlow: "debit",
										TransactionType: "bonus",
										Message:         &msg3,
										Status:          "Success",
									}
									tx.Create(&trx3)
								}
							}
						}
					}
				}
			}
			return nil
		})
		utils.WriteJSON(w, http.StatusOK, utils.APIResponse{Success: true, Message: "OK"})
		return
	}

	_ = db.Transaction(func(tx *gorm.DB) error {
		_ = tx.Model(&models.Transaction{}).Where("order_id = ?", inv.OrderID).Update("status", "Failed").Error
		_ = tx.Model(&inv).Update("status", "Cancelled").Error
		return nil
	})
	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{Success: true, Message: "Failed updated"})
}

// POST /api/cron/daily-returns
func CronDailyReturnsHandler(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("X-CRON-KEY")
	if key == "" || key != os.Getenv("CRON_KEY") {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Unauthorized"})
		return
	}

	db := database.DB
	now := time.Now()
	var due []models.Investment
	if err := db.Where("status = 'Running' AND next_return_at IS NOT NULL AND next_return_at <= ? AND total_paid < duration", now).Find(&due).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan"})
		return
	}
	processed := 0
	for i := range due {
		inv := due[i]
		_ = db.Transaction(func(tx *gorm.DB) error {
			var user models.User
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, inv.UserID).Error; err != nil {
				return err
			}

			amount := inv.DailyProfit
			newBalance := round3(user.Balance + amount)
			if err := tx.Model(&user).Update("balance", newBalance).Error; err != nil {
				return err
			}

			orderID := utils.GenerateOrderID(inv.UserID)
			msg := fmt.Sprintf("Daily profit investasi #%d", inv.ID)
			trx := models.Transaction{
				UserID:          inv.UserID,
				Amount:          amount,
				Charge:          0,
				OrderID:         orderID,
				TransactionFlow: "debit",
				TransactionType: "return",
				Message:         &msg,
				Status:          "Success",
			}
			if err := tx.Create(&trx).Error; err != nil {
				return err
			}

			levelPercents := []float64{0.05, 0.02, 0.01}
			levelMsgs := []string{
				"Bonus manajemen tim level 1",
				"Bonus manajemen tim level 2",
				"Bonus manajemen tim level 3",
			}
			currReffBy := user.ReffBy
			userLevel := user.Level
			for level := 0; level < 3 && currReffBy != nil; level++ {
				var reffUser models.User
				if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id, balance, reff_by, level").Where("id = ?", *currReffBy).First(&reffUser).Error; err != nil {
					break
				}
				var reffLevelVal, userLevelVal uint
				if rlPtr, ok := any(reffUser.Level).(*uint); ok && rlPtr != nil {
					reffLevelVal = *rlPtr
				} else {
					if reffUser.Level != nil {
						reffLevelVal = *reffUser.Level
					} else {
						reffLevelVal = 0
					}
				}
				if ulPtr, ok := any(userLevel).(*uint); ok && ulPtr != nil {
					userLevelVal = *ulPtr
				} else {
					if userLevel != nil {
						userLevelVal = *userLevel
					} else {
						userLevelVal = 0
					}
				}
				if reffLevelVal < userLevelVal {
					currReffBy = reffUser.ReffBy
					continue
				}
				bonus := round3(amount * levelPercents[level])
				if bonus > 0 {
					newReffBalance := round3(reffUser.Balance + bonus)
					if err := tx.Model(&reffUser).Update("balance", newReffBalance).Error; err != nil {
						return err
					}
					orderID := utils.GenerateOrderID(reffUser.ID)
					msg := levelMsgs[level]
					trxBonus := models.Transaction{
						UserID:          reffUser.ID,
						Amount:          bonus,
						Charge:          0,
						OrderID:         orderID,
						TransactionFlow: "debit",
						TransactionType: "team",
						Message:         &msg,
						Status:          "Success",
					}
					if err := tx.Create(&trxBonus).Error; err != nil {
						return err
					}
				}
				currReffBy = reffUser.ReffBy
			}

			now := time.Now()
			next := now.Add(24 * time.Hour)
			paid := inv.TotalPaid + 1
			returned := round3(inv.TotalReturned + amount)
			updates := map[string]interface{}{"total_paid": paid, "total_returned": returned, "last_return_at": now, "next_return_at": next}
			if paid >= inv.Duration {
				updates["status"] = "Completed"
			}
			if err := tx.Model(&inv).Updates(updates).Error; err != nil {
				return err
			}
			processed++
			return nil
		})
	}
	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{Success: true, Message: "Cron executed", Data: map[string]interface{}{"processed": processed}})
}

func parseTimeFlexible(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, errors.New("empty")
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006-01-02T15:04:05.000Z07:00", s); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("cannot parse time: %s", s)
}

func getKytaAccessToken(ctx context.Context, client *http.Client, baseURL, clientID, clientSecret string) (string, error) {
	url := strings.TrimRight(baseURL, "/") + "/access-token"

	credentials := clientID + ":" + clientSecret
	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(credentials))

	payload := map[string]string{
		"grant_type": "client_credentials",
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Basic "+encodedCredentials)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get access token, status: %d", resp.StatusCode)
	}

	var tokenResp KytaAccessTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	if tokenResp.ResponseData.AccessToken == "" {
		return "", errors.New("access token is empty")
	}

	return tokenResp.ResponseData.AccessToken, nil
}

func createKytaQRIS(ctx context.Context, client *http.Client, baseURL, accessToken, referenceID string, amount int64, notifyURL, successURL, failedURL string) (*KytaPaymentResponse, error) {
	url := strings.TrimRight(baseURL, "/") + "/payments/create/qris"

	payload := map[string]interface{}{
		"reference_id": referenceID,
		"amount":       amount,
		"notify_url":   notifyURL,
		"success_url":  successURL,
		"failed_url":   failedURL,
		"expires_time": 900,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody := new(bytes.Buffer)
		respBody.ReadFrom(resp.Body)
		return nil, fmt.Errorf("failed to create QRIS payment, status: %d, body: %s", resp.StatusCode, respBody.String())
	}

	var paymentResp KytaPaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&paymentResp); err != nil {
		return nil, err
	}

	return &paymentResp, nil
}

func createKytaVA(ctx context.Context, client *http.Client, baseURL, accessToken, referenceID string, amount int64, bankCode, notifyURL, successURL, failedURL string) (*KytaPaymentResponse, error) {
	url := strings.TrimRight(baseURL, "/") + "/payments/create/va"

	payload := map[string]interface{}{
		"reference_id": referenceID,
		"amount":       amount,
		"bank_code":    bankCode,
		"notify_url":   notifyURL,
		"success_url":  successURL,
		"failed_url":   failedURL,
		"expires_time": 900,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody := new(bytes.Buffer)
		respBody.ReadFrom(resp.Body)
		return nil, fmt.Errorf("failed to create VA payment, status: %d, body: %s", resp.StatusCode, respBody.String())
	}

	var paymentResp KytaPaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&paymentResp); err != nil {
		return nil, err
	}

	return &paymentResp, nil
}

func round3(f float64) float64 {
	return float64(int(f*100+0.5)) / 100
}
