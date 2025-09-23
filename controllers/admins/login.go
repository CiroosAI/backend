package admins

import (
	"encoding/json"
	"net/http"
	"project/models"
	"project/utils"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Invalid JSON body",
		})
		return
	}

	// Get admin by username
	admin, err := models.GetAdminByUsername(req.Username)
	if err != nil {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{
			Success: false,
			Message: "Username atau password salah",
		})
		return
	}

	// Validate password
	if !admin.ValidatePassword(req.Password) {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{
			Success: false,
			Message: "Username atau password salah",
		})
		return
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(admin.ID, admin.Username, "admin")
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal membuat token",
		})
		return
	}

	// Create successful response
	response := utils.APIResponse{
		Success: true,
		Message: "Berhasil login",
		Data: map[string]interface{}{
			"token": token,
			"admin": admin,
		},
	}

	utils.WriteJSON(w, http.StatusOK, response)
}
