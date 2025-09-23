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

type ProductResponse struct {
	ID         uint    `json:"id"`
	Name       string  `json:"name"`
	Minimum    float64 `json:"minimum"`
	Maximum    float64 `json:"maximum"`
	Percentage float64 `json:"percentage"`
	Duration   int     `json:"duration"`
	Status     string  `json:"status"`
}

func GetProducts(w http.ResponseWriter, r *http.Request) {
	var products []models.Product
	if err := database.DB.Find(&products).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Terjadi kesalahan sistem, silakan coba lagi",
		})
		return
	}

	var response []ProductResponse
	for _, product := range products {
		response = append(response, ProductResponse{
			ID:         product.ID,
			Name:       product.Name,
			Minimum:    product.Minimum,
			Maximum:    product.Maximum,
			Percentage: product.Percentage,
			Duration:   product.Duration,
			Status:     product.Status,
		})
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data:    response,
	})
}

type UpdateProductRequest struct {
	Name       string  `json:"name"`
	Minimum    float64 `json:"minimum"`
	Maximum    float64 `json:"maximum"`
	Percentage float64 `json:"percentage"`
	Duration   int     `json:"duration"`
	Status     string  `json:"status"`
}

func UpdateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "ID produk tidak valid",
		})
		return
	}

	// Only allow updating products with ID 1, 2, or 3
	if id > 3 {
		utils.WriteJSON(w, http.StatusForbidden, utils.APIResponse{
			Success: false,
			Message: "Hanya produk 1-3 yang dapat diupdate",
		})
		return
	}

	var req UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Format data tidak valid",
		})
		return
	}

	// Validate input
	if req.Minimum <= 0 || req.Maximum <= 0 || req.Percentage <= 0 || req.Duration <= 0 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Nilai minimum, maximum, percentage, dan duration harus lebih besar dari 0",
		})
		return
	}

	if req.Minimum >= req.Maximum {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Nilai minimum harus lebih kecil dari maximum",
		})
		return
	}

	if req.Status != "Active" && req.Status != "Inactive" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Status hanya boleh Active atau Inactive",
		})
		return
	}

	var product models.Product
	if err := database.DB.First(&product, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{
				Success: false,
				Message: "Produk tidak ditemukan",
			})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal mengambil data produk",
		})
		return
	}

	// Update fields
	product.Name = req.Name
	product.Minimum = req.Minimum
	product.Maximum = req.Maximum
	product.Percentage = req.Percentage
	product.Duration = req.Duration
	product.Status = req.Status

	if err := database.DB.Save(&product).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal memperbarui produk",
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Produk berhasil diupdate",
		Data: ProductResponse{
			ID:         product.ID,
			Name:       product.Name,
			Minimum:    product.Minimum,
			Maximum:    product.Maximum,
			Percentage: product.Percentage,
			Duration:   product.Duration,
			Status:     product.Status,
		},
	})
}
