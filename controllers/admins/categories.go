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

// GET /api/admin/categories
func ListCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	db := database.DB
	var categories []models.Category
	if err := db.Order("id ASC").Find(&categories).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal mengambil data kategori"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data: map[string]interface{}{
			"categories": categories,
		},
	})
}

// GET /api/admin/categories/{id}
func GetCategoryHandler(w http.ResponseWriter, r *http.Request) {
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
	var category models.Category
	if err := db.First(&category, uint(id64)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{Success: false, Message: "Kategori tidak ditemukan"})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data:    category,
	})
}

// POST /api/admin/categories
func CreateCategoryHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		ProfitType  string `json:"profit_type"`
		Status      string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Invalid JSON"})
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Nama kategori wajib diisi"})
		return
	}

	if req.ProfitType != "locked" && req.ProfitType != "unlocked" {
		req.ProfitType = "unlocked"
	}

	if req.Status != "Active" && req.Status != "Inactive" {
		req.Status = "Active"
	}

	category := models.Category{
		Name:        req.Name,
		Description: req.Description,
		ProfitType:  req.ProfitType,
		Status:      req.Status,
	}

	db := database.DB
	if err := db.Create(&category).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal membuat kategori"})
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.APIResponse{
		Success: true,
		Message: "Kategori berhasil dibuat",
		Data:    category,
	})
}

// PUT /api/admin/categories/{id}
func UpdateCategoryHandler(w http.ResponseWriter, r *http.Request) {
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
		Name        string `json:"name"`
		Description string `json:"description"`
		ProfitType  string `json:"profit_type"`
		Status      string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Invalid JSON"})
		return
	}

	db := database.DB
	var category models.Category
	if err := db.First(&category, uint(id64)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{Success: false, Message: "Kategori tidak ditemukan"})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan"})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.ProfitType == "locked" || req.ProfitType == "unlocked" {
		updates["profit_type"] = req.ProfitType
	}
	if req.Status == "Active" || req.Status == "Inactive" {
		updates["status"] = req.Status
	}

	if len(updates) > 0 {
		if err := db.Model(&category).Updates(updates).Error; err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal mengupdate kategori"})
			return
		}
	}

	// Reload to get updated data
	db.First(&category, uint(id64))

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Kategori berhasil diupdate",
		Data:    category,
	})
}

// DELETE /api/admin/categories/{id}
func DeleteCategoryHandler(w http.ResponseWriter, r *http.Request) {
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
	var category models.Category
	if err := db.First(&category, uint(id64)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{Success: false, Message: "Kategori tidak ditemukan"})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan"})
		return
	}

	// Check if any products use this category
	var count int64
	if err := db.Model(&models.Product{}).Where("category_id = ?", uint(id64)).Count(&count).Error; err == nil && count > 0 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Tidak dapat menghapus kategori yang masih digunakan oleh produk"})
		return
	}

	if err := db.Delete(&category).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal menghapus kategori"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Kategori berhasil dihapus",
	})
}

