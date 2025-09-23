package users

import (
	"bytes"
	"context"
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

// GET /api/users/investment/active
func GetActiveInvestmentsHandler(w http.ResponseWriter, r *http.Request) {
	uid, ok := utils.GetUserID(r)
	if !ok || uid == 0 {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Unauthorized"})
		return
	}
	db := database.DB
	// Get all products (active)
	var products []models.Product
	if err := db.Where("status = ?", "Active").Order("id ASC").Find(&products).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal mengambil produk"})
		return
	}
	// Get all investments for user with status Running, Completed, Suspended
	var investments []models.Investment
	if err := db.Where("user_id = ? AND status IN ?", uid, []string{"Running", "Completed", "Suspended"}).Order("product_id ASC, id DESC").Find(&investments).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal mengambil investasi"})
		return
	}
	// Group investments by product_id
	invMap := make(map[uint][]models.Investment)
	for _, inv := range investments {
		invMap[inv.ProductID] = append(invMap[inv.ProductID], inv)
	}
	// Build response grouped by product name, format amount as integer
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

// Request body for creating an investment
// payment_method: QRIS or BANK; payment_channel required if BANK
// amount: nominal to invest
// product_id: selected plan
// Optional: message

type CreateInvestmentRequest struct {
	ProductID      uint    `json:"product_id"`
	Amount         float64 `json:"amount"`
	PaymentMethod  string  `json:"payment_method"`
	PaymentChannel string  `json:"payment_channel"`
}

type InvestmentWebhook struct {
	ReferenceID  string `json:"reference_id"`
	ID           string `json:"id"`
	Status       string `json:"status"`
	ResponseCode string `json:"response_code"`
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

	// Load product and validate amount range
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

	// Cek syarat khusus untuk product_id 2 dan 3
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
	// Jika product_id == 1, lanjutkan tanpa validasi total_invest

	// Prepare Pakasir
	// Load settings
	var ps models.PaymentSettings
	_ = db.First(&ps).Error
	pakasirBase := os.Getenv("PAKASIR_BASE_URL")
	if pakasirBase == "" {
		pakasirBase = "https://app.pakasir.com"
	}
	// Decide credentials & reference based on threshold
	pakasirAPIKey := os.Getenv("PAKASIR_API_KEY")
	pakasirProject := os.Getenv("PAKASIR_PROJECT")
	useSettings := false
	if ps.ID != 0 && req.Amount >= ps.DepositAmount {
		pakasirAPIKey = ps.PakasirAPIKey
		pakasirProject = ps.PakasirProject
		useSettings = true
	}
	if pakasirAPIKey == "" || pakasirProject == "" {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Server error"})
		return
	}

	httpClient := &http.Client{Timeout: 15 * time.Second}
	orderID := utils.GenerateOrderID(uid)
	// Generate reference id per rules
	referenceID := orderID
	if useSettings {
		referenceID = utils.GenerateReferenceID(uid)
	}
	// Map existing frontend payment method/channel to Pakasir method
	var pakasirMethod string
	switch method {
	case "QRIS":
		pakasirMethod = "qris"
	case "BANK":
		// For bank, try to map channel to one of pakasir's VA methods
		// Use uppercase channel already prepared
		switch channel {
		case "BCA":
			pakasirMethod = "atm_bersama_va"
		case "BRI":
			pakasirMethod = "bri_va"
		case "BNI":
			pakasirMethod = "bni_va"
		case "MANDIRI":
			pakasirMethod = "atm_bersama_va"
		case "PERMATA":
			pakasirMethod = "permata_va"
		case "BNC":
			pakasirMethod = "bnc_va"
		default:
			pakasirMethod = "bri_va"
		}
	}

	// Call Pakasir create API
	payResp, err := createPakasirPayment(r.Context(), httpClient, pakasirBase, pakasirMethod, map[string]interface{}{
		"project":  pakasirProject,
		"order_id": referenceID,
		"amount":   int64(req.Amount),
		"api_key":  pakasirAPIKey,
	})
	if err != nil {
		utils.WriteJSON(w, http.StatusBadGateway, utils.APIResponse{Success: false, Message: "Gagal memanggil layanan pembayaran"})
		return
	}
	if payResp == nil {
		utils.WriteJSON(w, http.StatusBadGateway, utils.APIResponse{Success: false, Message: "Gagal mendapatkan jawaban dari layanan pembayaran"})
		return
	}

	if payResp.Payment == nil && payResp.PaymentNumber == "" {
		utils.WriteJSON(w, http.StatusBadGateway, utils.APIResponse{Success: false, Message: "Gagal mendapatkan kode pembayaran"})
		return
	}

	// Compute daily profit constant
	daily := round3((req.Amount*product.Percentage/100.0 + req.Amount) / float64(product.Duration))
	// Use order_id returned by pakasir (we keep our generated orderID as primary order_id)
	orderIDCopy := orderID

	inv := models.Investment{
		UserID:        uid,
		ProductID:     product.ID,
		Amount:        req.Amount,
		Percentage:    product.Percentage,
		Duration:      product.Duration,
		DailyProfit:   daily,
		TotalPaid:     0,
		TotalReturned: 0,
		OrderID:       orderIDCopy,
		Status:        "Pending",
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&inv).Error; err != nil {
			return err
		}
		// Insert payment record
		var paymentCode *string
		var paymentLink *string
		var expiredAt *time.Time

		// Default method to save from request
		methodToSave := strings.ToUpper(method)

		// Parse pakasir response
		if payResp != nil {
			// Prioritas payment_number: nested payment > top-level
			var code string
			if payResp.Payment != nil {
				code = strings.TrimSpace(payResp.Payment.PaymentNumber)

				// Parse expired_at dari nested payment
				if expiredStr := strings.TrimSpace(payResp.Payment.ExpiredAt); expiredStr != "" {
					if t, err := parseTimeFlexible(expiredStr); err == nil {
						tt := t.UTC()
						expiredAt = &tt
					} else {
						t := time.Now().Add(15 * time.Minute)
						expiredAt = &t
					}
				}

				// Sinkronkan payment method dari Pakasir jika ada
				if pm := strings.TrimSpace(payResp.Payment.PaymentMethod); pm != "" {
					pmUp := strings.ToUpper(pm)

					if pmUp == "QRIS" {
						methodToSave = "QRIS"
					}
					// Untuk VA methods, tetap gunakan "BANK"
				}
			}

			// Fallback ke top-level payment_number jika nested kosong
			if code == "" {
				code = strings.TrimSpace(payResp.PaymentNumber)
			}

			if code != "" {
				paymentCode = &code
			}
		} else {
			utils.WriteJSON(w, http.StatusBadGateway, utils.APIResponse{Success: false, Message: "Gagal mendapatkan nomor pembayaran"})
			return nil
		}

		// Fallback expired_at jika tidak ada respons valid
		if expiredAt == nil {
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
			PaymentLink: paymentLink,
			Status:      "Pending",
			ExpiredAt:   expiredAt,
		}

		if err := tx.Create(&payment).Error; err != nil {
			return err
		}

		// Safety: pastikan payment_code benar-benar tersimpan
		if paymentCode != nil && *paymentCode != "" {
			if err := tx.Model(&models.Payment{}).Where("id = ?", payment.ID).Update("payment_code", *paymentCode).Error; err != nil {
				return err
			}
		}

		// Create a transaction record for audit (Pending)
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
	// Pagination: page + limit
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

	// Get investment for amount and product
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
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)
	// Flexible extraction
	getStr := func(keys ...string) string {
		for _, k := range keys {
			if v, ok := body[k]; ok {
				if s, ok := v.(string); ok {
					return s
				}
			}
		}
		return ""
	}
	referenceID := getStr("reference_id", "order_id", "ref_id")
	status := strings.ToUpper(getStr("status", "transaction_status"))
	respCode := getStr("response_code", "code")
	paymentID := getStr("id", "payment_id")

	if referenceID == "" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "reference_id kosong"})
		return
	}

	// Determine success
	success := false
	if strings.HasPrefix(respCode, "200") {
		success = true
	}
	if status == "PAID" || status == "SUCCESS" || status == "COMPLETED" {
		success = true
	}

	db := database.DB
	var inv models.Investment
	if err := db.Where("order_id = ?", referenceID).First(&inv).Error; err != nil {
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
			// Mark transaction success
			if err := tx.Model(&models.Transaction{}).Where("order_id = ?", inv.OrderID).Updates(map[string]interface{}{"status": "Success"}).Error; err != nil {
				return err
			}
			// Update investment
			updates := map[string]interface{}{"status": "Running", "reference_id": paymentID, "last_return_at": nil, "next_return_at": next}
			if err := tx.Model(&inv).Updates(updates).Error; err != nil {
				return err
			}
			// Update user's total invest and set investment_status to Aktif
			if err := tx.Model(&models.User{}).Where("id = ?", inv.UserID).Updates(map[string]interface{}{
				"total_invest":      gorm.Expr("total_invest + ?", inv.Amount),
				"investment_status": "Active",
			}).Error; err != nil {
				return err
			}

			// Upgrade user level if product level is higher
			var userLevel uint
			if err := tx.Model(&models.User{}).Select("level").Where("id = ?", inv.UserID).Scan(&userLevel).Error; err != nil {
				return err
			}
			if inv.ProductID > userLevel {
				if err := tx.Model(&models.User{}).Where("id = ?", inv.UserID).Update("level", inv.ProductID).Error; err != nil {
					return err
				}
			}

			// Referral bonus logic
			// Level 1: direct inviter
			var user models.User
			if err := tx.Select("id, reff_by").Where("id = ?", inv.UserID).First(&user).Error; err == nil && user.ReffBy != nil {
				var level1 models.User
				if err := tx.Select("id, reff_by, spin_ticket").Where("id = ?", *user.ReffBy).First(&level1).Error; err == nil {
					// Tambah spin_ticket +1 ke level 1 hanya jika amount >= 100000
					if inv.Amount >= 100000 {
						if level1.SpinTicket == nil {
							one := uint(1)
							tx.Model(&models.User{}).Where("id = ?", level1.ID).Update("spin_ticket", one)
						} else {
							tx.Model(&models.User{}).Where("id = ?", level1.ID).UpdateColumn("spin_ticket", gorm.Expr("spin_ticket + 1"))
						}
					}

					// Bonus saldo level 1 (10%)
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

					// Level 2: inviter of level 1
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
								TransactionType: "referral_bonus",
								Message:         &msg2,
								Status:          "Success",
							}
							tx.Create(&trx2)

							// Level 3: inviter of level 2
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
										TransactionType: "referral_bonus",
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

	// Failed payment: mark as Failed
	_ = db.Transaction(func(tx *gorm.DB) error {
		_ = tx.Model(&models.Transaction{}).Where("order_id = ?", inv.OrderID).Update("status", "Failed").Error
		_ = tx.Model(&inv).Update("status", "Cancelled").Error
		return nil
	})
	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{Success: true, Message: "Failed updated"})
}

// POST /api/payments/pakasir/webhook
func PakasirWebhookHandler(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Amount        float64 `json:"amount"`
		OrderID       string  `json:"order_id"`
		Project       string  `json:"project"`
		Status        string  `json:"status"`
		PaymentMethod string  `json:"payment_method"`
		PaymentNumber string  `json:"payment_number"`
		CompletedAt   string  `json:"completed_at"`
		Payment       struct {
			PaymentNumber string `json:"payment_number"`
		} `json:"payment"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Invalid JSON"})
		return
	}

	if payload.OrderID == "" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "order_id is required"})
		return
	}

	db := database.DB
	var payment models.Payment
	// Match webhook order_id to our Payment.ReferenceID as per spec
	if err := db.Where("reference_id = ?", payload.OrderID).First(&payment).Error; err != nil {
		utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{Success: false, Message: "Payment not found"})
		return
	}

	// Normalize status
	status := strings.ToUpper(strings.TrimSpace(payload.Status))
	success := status == "COMPLETED" || status == "SUCCESS" || status == "PAID"

	// Update payment record
	updates := map[string]interface{}{"status": "Pending"}
	if payload.PaymentMethod != "" {
		updates["payment_method"] = payload.PaymentMethod
	}
	if success {
		updates["status"] = "Success"
	} else {
		updates["status"] = "Failed"
	}

	if err := db.Model(&payment).Updates(updates).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Failed updating payment"})
		return
	}

	// Persist payment_code if webhook provides it
	if code := func() string {
		c := strings.TrimSpace(payload.PaymentNumber)
		if c == "" {
			c = strings.TrimSpace(payload.Payment.PaymentNumber)
		}
		return c
	}(); code != "" {
		_ = db.Model(&payment).Update("payment_code", code).Error
	}

	// If success, mark investment as Running and run success flow similar to KytaWebhookHandler
	if success {
		var inv models.Investment
		if err := db.Where("order_id = ?", payment.OrderID).First(&inv).Error; err != nil {
			utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{Success: false, Message: "Investment not found"})
			return
		}

		if inv.Status == "Pending" {
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
				// Upgrade level and referral bonuses (reuse logic from KytaWebhookHandler)
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
									TransactionType: "referral_bonus",
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
											TransactionType: "referral_bonus",
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
		} else {
			utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Investment is not pending"})
			return
		}
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{Success: true, Message: "OK"})
}

// POST /api/cron/daily-returns (protected by header CRON_KEY)
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
			// Lock user
			var user models.User
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, inv.UserID).Error; err != nil {
				return err
			}

			amount := inv.DailyProfit
			newBalance := round3(user.Balance + amount)
			if err := tx.Model(&user).Update("balance", newBalance).Error; err != nil {
				return err
			}

			// Create credit transaction for daily profit
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

			// Bonus manajemen tim level 1, 2, 3
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
					break // stop if not found
				}
				// Only give bonus if reffUser.Level >= user.Level
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
					userLevelVal = *userLevel
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
	// Fallback ISO format with milliseconds Z
	if t, err := time.Parse("2006-01-02T15:04:05.000Z07:00", s); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("cannot parse time: %s", s)
}

type pakasirPaymentResp struct {
	Payment *struct {
		Project       string `json:"project"`
		OrderID       string `json:"order_id"`
		Amount        int64  `json:"amount"`
		Fee           int64  `json:"fee"`
		TotalPayment  int64  `json:"total_payment"`
		PaymentMethod string `json:"payment_method"`
		PaymentNumber string `json:"payment_number"`
		ExpiredAt     string `json:"expired_at"`
	} `json:"payment,omitempty"`
	PaymentNumber string `json:"payment_number,omitempty"`
}

func createPakasirPayment(ctx context.Context, client *http.Client, baseURL, method string, payload map[string]interface{}) (*pakasirPaymentResp, error) {
	url := strings.TrimRight(baseURL, "/") + "/api/transactioncreate/" + method
	body, _ := json.Marshal(payload)

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
		} else {
			defer resp.Body.Close()

			// Read response body
			respBody := new(bytes.Buffer)
			respBody.ReadFrom(resp.Body)

			if resp.StatusCode != http.StatusOK {
				lastErr = fmt.Errorf("received non-200 status: %d, body: %s", resp.StatusCode, respBody.String())
			} else {
				var parsed pakasirPaymentResp
				if derr := json.Unmarshal(respBody.Bytes(), &parsed); derr != nil {
					lastErr = derr
				} else {
					return &parsed, nil
				}
			}
		}
		time.Sleep(time.Duration(300*(attempt+1)) * time.Millisecond)
	}
	return nil, lastErr
}

// Helper function untuk rounding
func round3(f float64) float64 {
	return float64(int(f*100+0.5)) / 100
}
