package admins

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"project/database"
	"project/models"
	"project/utils"

	"gorm.io/gorm"
)

// GET /api/admin/products
func ListProductsHandler(w http.ResponseWriter, r *http.Request) {
	db := database.DB
	var products []models.Product
	if err := db.Preload("Category").Order("category_id ASC, id ASC").Find(&products).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal mengambil data produk"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data: map[string]interface{}{
			"products": products,
		},
	})
}

// GET /api/admin/products/{id}
func GetProductHandler(w http.ResponseWriter, r *http.Request) {
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
	var product models.Product
	if err := db.Preload("Category").First(&product, uint(id64)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{Success: false, Message: "Produk tidak ditemukan"})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data:    product,
	})
}

// POST /api/admin/products
func CreateProductHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CategoryID    uint    `json:"category_id"`
		Name          string  `json:"name"`
		Amount        float64 `json:"amount"`
		DailyProfit   float64 `json:"daily_profit"`
		Duration      int     `json:"duration"`
		RequiredVIP   int     `json:"required_vip"`
		PurchaseLimit int     `json:"purchase_limit"`
		Status        string  `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Invalid JSON"})
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Nama produk wajib diisi"})
		return
	}

	if req.CategoryID == 0 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Kategori wajib dipilih"})
		return
	}

	if req.Amount <= 0 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Amount harus lebih dari 0"})
		return
	}

	if req.DailyProfit <= 0 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Daily profit harus lebih dari 0"})
		return
	}

	if req.Duration <= 0 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Duration harus lebih dari 0"})
		return
	}

	if req.Status != "Active" && req.Status != "Inactive" {
		req.Status = "Active"
	}

	db := database.DB

	// Check if category exists
	var category models.Category
	if err := db.First(&category, req.CategoryID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Kategori tidak ditemukan"})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan"})
		return
	}

	product := models.Product{
		CategoryID:    req.CategoryID,
		Name:          req.Name,
		Amount:        req.Amount,
		DailyProfit:   req.DailyProfit,
		Duration:      req.Duration,
		RequiredVIP:   req.RequiredVIP,
		PurchaseLimit: req.PurchaseLimit,
		Status:        req.Status,
	}

	if err := db.Create(&product).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal membuat produk"})
		return
	}

	// Reload with category
	db.Preload("Category").First(&product, product.ID)

	utils.WriteJSON(w, http.StatusCreated, utils.APIResponse{
		Success: true,
		Message: "Produk berhasil dibuat",
		Data:    product,
	})
}

// PUT /api/admin/products/{id}
func UpdateProductHandler(w http.ResponseWriter, r *http.Request) {
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

	var req struct {
		CategoryID    *uint    `json:"category_id"`
		Name          string   `json:"name"`
		Amount        *float64 `json:"amount"`
		DailyProfit   *float64 `json:"daily_profit"`
		Duration      *int     `json:"duration"`
		RequiredVIP   *int     `json:"required_vip"`
		PurchaseLimit *int     `json:"purchase_limit"`
		Status        string   `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Invalid JSON"})
		return
	}

	db := database.DB
	var product models.Product
	if err := db.First(&product, uint(id64)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{Success: false, Message: "Produk tidak ditemukan"})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan"})
		return
	}

	updates := map[string]interface{}{}
	
	if req.CategoryID != nil && *req.CategoryID > 0 {
		// Check if category exists
		var category models.Category
		if err := db.First(&category, *req.CategoryID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Kategori tidak ditemukan"})
				return
			}
		}
		updates["category_id"] = *req.CategoryID
	}
	
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Amount != nil && *req.Amount > 0 {
		updates["amount"] = *req.Amount
	}
	if req.DailyProfit != nil && *req.DailyProfit > 0 {
		updates["daily_profit"] = *req.DailyProfit
	}
	if req.Duration != nil && *req.Duration > 0 {
		updates["duration"] = *req.Duration
	}
	if req.RequiredVIP != nil {
		updates["required_vip"] = *req.RequiredVIP
	}
	if req.PurchaseLimit != nil {
		updates["purchase_limit"] = *req.PurchaseLimit
	}
	if req.Status == "Active" || req.Status == "Inactive" {
		updates["status"] = req.Status
	}

	if len(updates) > 0 {
		if err := db.Model(&product).Updates(updates).Error; err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal mengupdate produk"})
			return
		}
	}

	// Reload to get updated data
	db.Preload("Category").First(&product, uint(id64))

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Produk berhasil diupdate",
		Data:    product,
	})
}

// DELETE /api/admin/products/{id}
func DeleteProductHandler(w http.ResponseWriter, r *http.Request) {
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
	var product models.Product
	if err := db.First(&product, uint(id64)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{Success: false, Message: "Produk tidak ditemukan"})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan"})
		return
	}

	// Check if any investments use this product
	var count int64
	if err := db.Model(&models.Investment{}).Where("product_id = ?", uint(id64)).Count(&count).Error; err == nil && count > 0 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Tidak dapat menghapus produk yang masih digunakan oleh investasi"})
		return
	}

	if err := db.Delete(&product).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal menghapus produk"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Produk berhasil dihapus",
	})
}
