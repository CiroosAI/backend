package controllers

import (
	"net/http"

	"project/database"
	"project/models"
	"project/utils"
)

func ProductListHandler(w http.ResponseWriter, r *http.Request) {
	db := database.DB
	var products []models.Product
	if err := db.Where("status = ?", "Active").Order("id ASC").Find(&products).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan sistem, silakan coba lagi"})
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
